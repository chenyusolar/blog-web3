package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
	AI  *AIService
)

const BLOG_CACHE_KEY = "aigen_blog_list"
var JWT_SECRET = []byte("aigen_blog_secret_key")

func initDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	dsn := os.Getenv("DB_DSN")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// 自动迁移所有模型
	DB.AutoMigrate(&User{}, &Blog{}, &Comment{}, &Media{}, &Category{}, &Tag{})
	
	// 为存量数据补全邀请码和钱包
	backfillUserData()

	// 自动创建默认管理员账号
	seedAdmin()

	RDB = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	AI = NewAIService(
		os.Getenv("AI_API_KEY"),
		os.Getenv("AI_BASE_URL"),
		os.Getenv("AI_MODEL"),
	)
	
	// 创建上传目录
	os.MkdirAll("uploads", os.ModePerm)
}

// AuthMiddleware JWT 鉴权中间件
func AuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tokenStr := string(c.GetHeader("Authorization"))
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.H{"error": "Unauthorized"})
			return
		}

		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := uint(claims["user_id"].(float64))
			c.Set("user_id", userID)
			c.Next(ctx)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.H{"error": "Invalid Token"})
		}
	}
}

func main() {
	initDB()

	h := server.Default(server.WithHostPorts(fmt.Sprintf(":%s", os.Getenv("PORT"))))

	// 跨域与静态资源
	h.Use(corsMiddleware())
	h.StaticFS("/uploads", &app.FS{Root: "./uploads", PathRewrite: app.NewPathSlashesStripper(1)})

	v1 := h.Group("/api/v1")
	{
		v1.POST("/register", register)
		v1.POST("/login", login)
		v1.GET("/blogs", listBlogs)
		v1.GET("/blogs/:id", getBlogDetail)
		v1.GET("/categories", listCategories) // 新增分类接口
		v1.GET("/tags", listTags)             // 新增标签接口

		auth := v1.Group("/", AuthMiddleware())
		{
			auth.GET("/user/profile", getUserProfile)
			auth.PUT("/user/profile", updateUserProfile)
			auth.PUT("/user/password", changePassword)
			auth.GET("/user/blogs", getUserBlogs)
			auth.GET("/user/wallet", getWalletInfo)   // Web3 Wallet
			auth.GET("/user/rewards", getRewardLogs) // Reward History
			
			// Admin Routes
			admin := auth.Group("/admin")
			{
				admin.GET("/users", adminGetUsers)
				admin.GET("/stats", adminGetStats)
				admin.GET("/config", adminGetConfig)
				admin.PUT("/config", adminUpdateConfig)
			}
			
			auth.POST("/blogs/generate", generateBlog)
			auth.POST("/blogs", publishBlog)
			auth.POST("/blogs/:id/forward", forwardBlog) // Repost
			auth.PUT("/blogs/:id", updateBlog)
			auth.DELETE("/blogs/:id", deleteBlog)
			auth.POST("/comments", addComment)
			auth.POST("/upload", uploadFile)
			auth.POST("/categories", createCategory) // 管理分类
			auth.POST("/tags", createTag)             // 管理标签
		}
	}

	h.Spin()
}

// 跨域处理
func corsMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}
		c.Next(ctx)
	}
}

// 用户模块逻辑 (注册与登录)
func register(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Email        string `json:"email"`
		InviterCode  string `json:"inviter_code"`
		ImportWallet string `json:"import_wallet"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	// 查找邀请人
	var referrer User
	if req.InviterCode != "" {
		DB.Where("referral_code = ?", req.InviterCode).First(&referrer)
	}

	// 生成/设置钱包
	address := req.ImportWallet
	privateKey := ""
	if address == "" {
		var err error
		address, privateKey, err = GenerateSolanaWallet()
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{"error": "钱包生成失败"})
			return
		}
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := User{
		Username:      req.Username,
		PasswordHash:  string(hash),
		Email:         req.Email,
		WalletAddress: address,
		ReferralCode:  uuid.New().String()[:8],
	}

	if referrer.ID != 0 {
		user.ReferrerID = &referrer.ID
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, utils.H{"error": "用户已存在"})
		return
	}

	// 处理邀请奖励
	if user.ReferrerID != nil {
		go ProcessReferral(user.ID, user.ReferrerID)
	}

	// 发送私钥邮件
	if privateKey != "" {
		go SendWalletEmail(user.Email, user.Username, address, privateKey)
	}

	c.JSON(http.StatusOK, utils.H{"message": "注册成功"})
}

func login(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	var user User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "用户未找到"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "密码错误"})
		return
	}

	// 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenStr, _ := token.SignedString(JWT_SECRET)

	c.JSON(http.StatusOK, utils.H{"token": tokenStr, "user": user})
}

// 博客管理逻辑
func listBlogs(ctx context.Context, c *app.RequestContext) {
	var blogs []Blog
	query := DB.Preload("Author").Preload("Category").Preload("Tags")
	
	// 支持按分类过滤
	if catID := c.Query("category_id"); catID != "" {
		query = query.Where("category_id = ?", catID)
	}

	query.Order("created_at desc").Find(&blogs)
	c.JSON(http.StatusOK, blogs)
}

func getBlogDetail(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var blog Blog
	if err := DB.Preload("Author").Preload("Category").Preload("Tags").First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "博文未找到"})
		return
	}

	var comments []Comment
	DB.Preload("User").Where("blog_id = ?", id).Find(&comments)

	c.JSON(http.StatusOK, utils.H{"blog": blog, "comments": comments})
}

func publishBlog(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	
	var req struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		ImageURL   string `json:"image_url"`
		VideoURL   string `json:"video_url"`
		CategoryID uint   `json:"category_id"`
		TagNames   []string `json:"tag_names"`
	}
	
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "无效参数"})
		return
	}

	// 处理标签：如果不存在则创建
	var tags []Tag
	for _, name := range req.TagNames {
		var tag Tag
		DB.FirstOrCreate(&tag, Tag{Name: name})
		tags = append(tags, tag)
	}

	blog := Blog{
		Title:      req.Title,
		Content:    req.Content,
		ImageURL:   req.ImageURL,
		VideoURL:   req.VideoURL,
		AuthorID:   userID,
		CategoryID: &req.CategoryID,
		Tags:       tags,
	}

	DB.Create(&blog)
	RDB.Del(ctx, BLOG_CACHE_KEY)
	
	// 发放奖励
	go DistributeRewards(userID, "POST")
	
	c.JSON(http.StatusOK, blog)
}

// 分类与标签管理接口实现
func listCategories(ctx context.Context, c *app.RequestContext) {
	var cats []Category
	DB.Find(&cats)
	c.JSON(http.StatusOK, cats)
}

func createCategory(ctx context.Context, c *app.RequestContext) {
	var cat Category
	c.BindAndValidate(&cat)
	DB.Create(&cat)
	c.JSON(http.StatusOK, cat)
}

func listTags(ctx context.Context, c *app.RequestContext) {
	var tags []Tag
	DB.Find(&tags)
	c.JSON(http.StatusOK, tags)
}

func createTag(ctx context.Context, c *app.RequestContext) {
	var tag Tag
	c.BindAndValidate(&tag)
	DB.Create(&tag)
	c.JSON(http.StatusOK, tag)
}

// seedAdmin 自动创建初始数据 (管理员账号和一些默认分类)
func seedAdmin() {
	var count int64
	DB.Model(&User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		admin := User{
			Username:     "admin",
			PasswordHash: string(hash),
			Email:        "admin@example.com",
			Role:         "admin",
		}
		DB.Create(&admin)
		log.Println("Default admin account created: admin / admin")
	}

	// 同时也创建一些初始分类，方便系统立刻运行
	var catCount int64
	DB.Model(&Category{}).Count(&catCount)
	if catCount == 0 {
		defaultCats := []Category{
			{Name: "人工智能", Description: "AI 领域的新闻与洞察"},
			{Name: "编程技术", Description: "全栈开发与技术分享"},
			{Name: "生活杂记", Description: "博主的日常生活分享"},
		}
		for _, cat := range defaultCats {
			DB.Create(&cat)
		}
		log.Println("Default categories created.")
	}
}

func generateBlog(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var req struct {
		Prompt string `json:"prompt"`
	}
	c.BindAndValidate(&req)

	blog, err := AI.Generate(ctx, req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "AI 生成失败"})
		return
	}
	blog.AuthorID = userID
	DB.Create(blog)
	RDB.Del(ctx, BLOG_CACHE_KEY)
	c.JSON(http.StatusOK, blog)
}

func updateBlog(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	userID := c.GetUint("user_id")
	var blog Blog
	DB.First(&blog, id)
	if blog.AuthorID != userID {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权编辑他人文章"})
		return
	}
	c.BindAndValidate(&blog)
	DB.Save(&blog)
	c.JSON(http.StatusOK, blog)
}

func deleteBlog(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	DB.Delete(&Blog{}, id)
	c.JSON(http.StatusOK, utils.H{"message": "删除成功"})
}

// 评论功能
func addComment(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var comment Comment
	c.BindAndValidate(&comment)
	comment.UserID = userID
	DB.Create(&comment)
	
	// 发放奖励
	go DistributeRewards(userID, "COMMENT")
	
	c.JSON(http.StatusOK, comment)
}

// 文件上传管理
func uploadFile(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	fileHeader, _ := c.FormFile("file")
	
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext
	savePath := filepath.Join("uploads", newFileName)

	if err := c.SaveUploadedFile(fileHeader, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "上传失败"})
		return
	}

	media := Media{
		FileName:   fileHeader.Filename,
		FileURL:    "/uploads/" + newFileName,
		FileType:   fileHeader.Header.Get("Content-Type"),
		UploaderID: userID,
	}
	DB.Create(&media)

	c.JSON(http.StatusOK, media)
}
// 用户个人资料管理
func getUserProfile(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var user User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "用户未找到"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func updateUserProfile(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	var user User
	DB.First(&user, userID)
	
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusConflict, utils.H{"error": "更新失败，用户名或邮箱可能已存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func changePassword(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	var user User
	DB.First(&user, userID)

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, utils.H{"error": "旧密码错误"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	DB.Save(&user)

	c.JSON(http.StatusOK, utils.H{"message": "密码修改成功"})
}

func getUserBlogs(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	blogs := []Blog{}
	DB.Preload("Category").Preload("Tags").Where("author_id = ?", userID).Order("created_at desc").Find(&blogs)
	c.JSON(http.StatusOK, blogs)
}

func getWalletInfo(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var user User
	DB.First(&user, userID)
	c.JSON(http.StatusOK, utils.H{
		"address":       user.WalletAddress,
		"balance":       user.BlogBalance,
		"referral_code": user.ReferralCode,
	})
}

func getRewardLogs(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var logs []RewardLog
	DB.Where("user_id = ?", userID).Order("created_at desc").Find(&logs)
	c.JSON(http.StatusOK, logs)
}

func forwardBlog(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	blogID := c.Param("id")
	
	var original Blog
	if err := DB.First(&original, blogID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "原文未找到"})
		return
	}

	newBlog := Blog{
		Title:      "转发: " + original.Title,
		Content:    original.Content,
		AuthorID:   userID,
		IsForward:  true,
		OriginalID: &original.ID,
		CategoryID: original.CategoryID,
	}
	
	DB.Create(&newBlog)
	go DistributeRewards(userID, "FORWARD")
	
	c.JSON(http.StatusOK, newBlog)
}

// Admin 面板逻辑
func adminGetUsers(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var admin User
	DB.First(&admin, userID)
	if admin.Role != "admin" {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权访问此页面"})
		return
	}

	var users []User
	DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func adminGetStats(ctx context.Context, c *app.RequestContext) {
	var userCount, blogCount, commentCount int64
	var totalIssued float64
	DB.Model(&User{}).Count(&userCount)
	DB.Model(&Blog{}).Count(&blogCount)
	DB.Model(&Comment{}).Count(&commentCount)
	DB.Model(&RewardLog{}).Select("sum(amount)").Scan(&totalIssued)

	c.JSON(http.StatusOK, utils.H{
		"user_count":    userCount,
		"blog_count":    blogCount,
		"comment_count": commentCount,
		"total_issued":  totalIssued,
	})
}

func adminGetConfig(ctx context.Context, c *app.RequestContext) {
	// 这里读取 .env 信息 (实际生产中应更小心处理私钥)
	c.JSON(http.StatusOK, utils.H{
		"solana_rpc":     os.Getenv("SOLANA_RPC"),
		"admin_address":  os.Getenv("ADMIN_WALLET_ADDRESS"),
		"admin_key":      os.Getenv("ADMIN_WALLET_KEY"),
		"reward_post":    os.Getenv("REWARD_POST"),
		"reward_forward": os.Getenv("REWARD_FORWARD"),
		"reward_comment": os.Getenv("REWARD_COMMENT"),
		"ref_l1":         os.Getenv("REF_L1"),
		"ref_l2":         os.Getenv("REF_L2"),
		"ref_l3":         os.Getenv("REF_L3"),
	})
}

func adminUpdateConfig(ctx context.Context, c *app.RequestContext) {
	var req map[string]string
	c.BindAndValidate(&req)

	// 更新内存中的环境变量 (持久化需要额外逻辑)
	for k, v := range req {
		os.Setenv(k, v)
	}

	// 这里可以添加持久化到 .env 文件的逻辑
	c.JSON(http.StatusOK, utils.H{"message": "配置更新成功，请注意这些只是临时设置，重启失效"})
}

func backfillUserData() {
	var users []User
	DB.Find(&users)
	for _, user := range users {
		updated := false
		if user.ReferralCode == "" {
			user.ReferralCode = strings.ToUpper(uuid.New().String()[:8])
			updated = true
		}
		if user.WalletAddress == "" {
			addr, priv, _ := GenerateSolanaWallet()
			user.WalletAddress = addr
			log.Printf("Generated wallet for existing user %v: %v", user.Username, addr)
			SendWalletEmail(user.Email, user.Username, addr, priv)
			updated = true
		}
		if updated {
			DB.Save(&user)
		}
	}
}
