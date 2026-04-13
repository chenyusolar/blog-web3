package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	DB.AutoMigrate(&User{}, &Blog{}, &Comment{}, &Media{}, &Category{}, &Tag{}, &RewardLog{}, &VipApplication{})

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

		// Handle "Bearer <token>" format
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})

		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.H{"error": "Invalid or expired token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID := uint(claims["user_id"].(float64))
			c.Set("user_id", userID)
			c.Next(ctx)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.H{"error": "Invalid token claims"})
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
		v1.GET("/blogs/hot", hotBlogs)
		v1.GET("/blogs/vip", vipBlogs)
		v1.GET("/tags/hot", hotTags)
		v1.GET("/blogs/:id", getBlogDetail)
		v1.GET("/categories", listCategories) // 新增分类接口
		v1.GET("/tags", listTags)             // 新增标签接口

		auth := v1.Group("/", AuthMiddleware())
		{
			auth.GET("/user/profile", getUserProfile)
			auth.PUT("/user/profile", updateUserProfile)
			auth.PUT("/user/password", changePassword)
			auth.GET("/user/blogs", getUserBlogs)
			auth.GET("/user/wallet", getWalletInfo)  // Web3 Wallet
			auth.GET("/user/rewards", getRewardLogs) // Reward History
			auth.GET("/leaderboard", getLeaderboard) // Token Leaderboard
			auth.POST("/vip/apply", applyVip)        // Apply for VIP
			auth.GET("/vip/status", getVipStatus)    // Check VIP status
			auth.POST("/user/buy", buyTokens)        // Buy BLOG tokens

			// Admin Routes
			admin := auth.Group("/admin")
			{
				admin.GET("/users", adminGetUsers)
				admin.GET("/stats", adminGetStats)
				admin.GET("/config", adminGetConfig)
				admin.PUT("/config", adminUpdateConfig)
				admin.PUT("/users/:id/vip", adminToggleVip)
				admin.GET("/vip/applications", adminGetVipApplications)
				admin.PUT("/vip/applications/:id/approve", adminApproveVip)
				admin.PUT("/vip/applications/:id/reject", adminRejectVip)
			}

			auth.POST("/blogs/generate", generateBlog)
			auth.POST("/blogs", publishBlog)
			auth.POST("/blogs/:id/share", shareBlog) // External Social Share
			auth.PUT("/blogs/:id", updateBlog)
			auth.DELETE("/blogs/:id", deleteBlog)
			auth.POST("/comments", addComment)
			auth.POST("/upload", uploadFile)
			auth.POST("/categories", createCategory) // 管理分类
			auth.POST("/tags", createTag)            // 管理标签
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
	blogs := []Blog{}
	query := DB.Preload("Author").Preload("Category").Preload("Tags")

	// 支持按分类过滤
	if catID := c.Query("category_id"); catID != "" {
		query = query.Where("category_id = ?", catID)
	}

	var total int64
	query.Model(&Blog{}).Count(&total)

	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	page := 1
	pageSize := 18

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
		pageSize = ps
	}

	offset := (page - 1) * pageSize
	query.Preload("OriginalBlog").Preload("OriginalBlog.Author").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&blogs)

	c.JSON(http.StatusOK, utils.H{
		"blogs":     blogs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func hotBlogs(ctx context.Context, c *app.RequestContext) {
	blogs := []Blog{}
	DB.Preload("Author").Preload("Category").Preload("Tags").Order("view_count desc").Limit(10).Find(&blogs)
	c.JSON(http.StatusOK, blogs)
}

func vipBlogs(ctx context.Context, c *app.RequestContext) {
	var user User
	tokenStr := string(c.GetHeader("Authorization"))
	if tokenStr != "" {
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})
		if err == nil && token != nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				userID := uint(claims["user_id"].(float64))
				DB.First(&user, userID)
			}
		}
	}

	if user.ID == 0 || (!user.IsVip && user.Role != "admin") {
		c.JSON(http.StatusForbidden, utils.H{"error": "需要会员权限"})
		return
	}

	blogs := []Blog{}
	var total int64
	query := DB.Model(&Blog{}).Where("is_vip = ?", true)
	query.Count(&total)

	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")
	page := 1
	pageSize := 18
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
		pageSize = ps
	}

	offset := (page - 1) * pageSize
	DB.Preload("Author").Preload("Category").Preload("Tags").
		Where("is_vip = ?", true).
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&blogs)

	c.JSON(http.StatusOK, utils.H{
		"blogs":     blogs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func hotTags(ctx context.Context, c *app.RequestContext) {
	type HotTag struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	var hotTags []HotTag
	DB.Raw(`
		SELECT t.id, t.name, COALESCE(COUNT(bt.blog_id), 0) as count
		FROM tags t
		LEFT JOIN blog_tags bt ON t.id = bt.tag_id
		GROUP BY t.id, t.name
		ORDER BY count DESC
		LIMIT 15
	`).Scan(&hotTags)
	c.JSON(http.StatusOK, hotTags)
}

func getBlogDetail(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	var blog Blog
	if err := DB.Preload("Author").Preload("Category").Preload("Tags").Preload("OriginalBlog").Preload("OriginalBlog.Author").First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "博文未找到"})
		return
	}

	DB.Model(&blog).Update("view_count", gorm.Expr("view_count + 1"))

	var comments []Comment
	DB.Preload("User").Where("blog_id = ?", id).Find(&comments)

	tokenStr := string(c.GetHeader("Authorization"))
	isLocked := false
	if blog.IsVip && tokenStr != "" {
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})
		if err == nil && token != nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				userID := uint(claims["user_id"].(float64))
				var user User
				DB.First(&user, userID)
				if !user.IsVip && user.Role != "admin" {
					isLocked = true
				}
			}
		}
	} else if blog.IsVip {
		isLocked = true
	}

	if isLocked {
		blog.Content = "会员专享内容，请升级会员后查看"
	}

	c.JSON(http.StatusOK, utils.H{"blog": blog, "comments": comments, "is_locked": isLocked})
}

func publishBlog(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")

	var req struct {
		Title      string   `json:"title"`
		Content    string   `json:"content"`
		ImageURL   string   `json:"image_url"`
		VideoURL   string   `json:"video_url"`
		CategoryID uint     `json:"category_id"`
		TagNames   []string `json:"tag_names"`
		IsVip      bool     `json:"is_vip"`
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
		IsVip:      req.IsVip,
	}

	DB.Create(&blog)
	RDB.Del(ctx, BLOG_CACHE_KEY)

	// 发放奖励
	go DistributeRewards(userID, "POST")

	c.JSON(http.StatusOK, blog)
}

// 分类与标签管理接口实现
func listCategories(ctx context.Context, c *app.RequestContext) {
	cats := []Category{}
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
	tags := []Tag{}
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

	var tagCount int64
	DB.Model(&Tag{}).Count(&tagCount)
	if tagCount == 0 {
		defaultTags := []Tag{
			{Name: "AI"}, {Name: "GPT"}, {Name: "机器学习"}, {Name: "深度学习"},
			{Name: "Vue3"}, {Name: "Go语言"}, {Name: "区块链"}, {Name: "Web3"},
			{Name: "前端"}, {Name: "后端"}, {Name: "数据库"}, {Name: "Docker"},
		}
		for _, tag := range defaultTags {
			DB.Create(&tag)
		}
		log.Println("Default tags created.")
	}

	var vipBlogCount int64
	DB.Model(&Blog{}).Where("is_vip = ?", true).Count(&vipBlogCount)
	if vipBlogCount == 0 {
		var adminID uint
		DB.Raw("SELECT id FROM users WHERE role = 'admin' LIMIT 1").Scan(&adminID)
		var catID uint
		DB.Raw("SELECT id FROM categories WHERE name = '人工智能' LIMIT 1").Scan(&catID)

		vipBlogs := []Blog{
			{
				Title:      "【会员专享】2026年AI大模型最新突破：从GPT-5到AGI的距离",
				Content:    "<p>2026年，AI领域迎来了前所未有的突破。本文将深入分析GPT-5、Claude 4等最新大模型的技术细节，以及它们与AGI（通用人工智能）之间的距离。</p><h2>一、GPT-5 的核心突破</h2><p>GPT-5 在参数量、推理能力和多模态理解方面实现了质的飞跃...</p><h2>二、多模态融合的新范式</h2><p>最新模型已经能够同时处理文本、图像、音频和视频数据...</p><h2>三、AGI 的时间表</h2><p>业内专家预测，我们距离真正的AGI可能还有5-10年的时间...</p>",
				ImageURL:   "https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800",
				AuthorID:   adminID,
				IsVip:      true,
				CategoryID: &catID,
			},
			{
				Title:      "【会员专享】区块链DeFi投资策略：2026年收益最大化指南",
				Content:    "<p>去中心化金融（DeFi）在2026年已经发展成为一个万亿级别的市场。本文将为会员分享最新的DeFi投资策略和收益优化方法。</p><h2>一、流动性挖矿进阶</h2><p>通过多协议组合策略，可以在控制风险的同时获得更高的年化收益...</p><h2>二、跨链套利机会</h2><p>随着多链生态的成熟，跨链套利成为新的收益来源...</p><h2>三、风险管理框架</h2><p>建立完善的DeFi投资组合风险管理体系是长期盈利的关键...</p>",
				ImageURL:   "https://images.unsplash.com/photo-1639762681485-074b7f938ba0?w=800",
				AuthorID:   adminID,
				IsVip:      true,
				CategoryID: &catID,
			},
			{
				Title:      "【会员专享】Web3社交协议深度解析：去中心化社交的未来",
				Content:    "<p>Web3社交协议正在重塑社交媒体的格局。本文深入分析Lens Protocol、Farcaster等主流去中心化社交协议的技术架构和商业模式。</p><h2>一、Lens Protocol 生态系统</h2><p>Lens Protocol 通过NFT化的社交关系图谱，让用户真正拥有自己的社交数据...</p><h2>二、Farcaster 的创新设计</h2><p>Farcaster 采用混合架构，在去中心化和用户体验之间找到了平衡点...</p><h2>三、社交代币经济</h2><p>创作者经济在Web3时代迎来了新的可能性...</p>",
				ImageURL:   "https://images.unsplash.com/photo-1639322537228-f710d846310a?w=800",
				AuthorID:   adminID,
				IsVip:      true,
				CategoryID: &catID,
			},
			{
				Title:      "【会员专享】AI编程助手实战：从代码生成到智能架构设计",
				Content:    "<p>AI编程助手已经从简单的代码补全工具进化为能够进行架构设计的智能伙伴。本文分享高级使用技巧和最佳实践。</p><h2>一、提示词工程进阶</h2><p>掌握高级提示词设计方法，让AI生成更高质量的代码...</p><h2>二、架构设计辅助</h2><p>利用AI进行系统架构设计、技术选型和性能优化...</p><h2>三、团队协作工作流</h2><p>将AI编程助手集成到团队开发流程中，提升整体开发效率...</p>",
				ImageURL:   "https://images.unsplash.com/photo-1555949963-aa79dcee981c?w=800",
				AuthorID:   adminID,
				IsVip:      true,
				CategoryID: &catID,
			},
			{
				Title:      "【会员专享】Solana生态开发实战：从智能合约到DApp部署",
				Content:    "<p>Solana 作为高性能区块链平台，正在吸引越来越多的开发者。本文将从零开始教你在Solana生态中进行开发。</p><h2>一、Rust与Solana开发基础</h2><p>掌握Solana程序开发的核心概念和Rust编程基础...</p><h2>二、智能合约开发实战</h2><p>通过实际案例学习Solana智能合约的开发模式和最佳实践...</p><h2>三、前端DApp集成</h2><p>使用@solana/web3.js和钱包适配器构建完整的去中心化应用...</p>",
				ImageURL:   "https://images.unsplash.com/photo-1642104704074-907c0698cbd9?w=800",
				AuthorID:   adminID,
				IsVip:      true,
				CategoryID: &catID,
			},
		}

		for _, blog := range vipBlogs {
			DB.Create(&blog)
		}
		log.Println("Default VIP blogs created.")
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

	var currentUser User
	DB.First(&currentUser, userID)

	var blog Blog
	if err := DB.First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "文章未找到"})
		return
	}

	if blog.AuthorID != userID && currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权编辑他人文章"})
		return
	}

	var req Blog
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "无效参数"})
		return
	}

	// 只更新部分字段，避免覆盖 ID 或 AuthorID
	DB.Model(&blog).Updates(map[string]interface{}{
		"title":       req.Title,
		"content":     req.Content,
		"image_url":   req.ImageURL,
		"video_url":   req.VideoURL,
		"category_id": req.CategoryID,
		"is_vip":      req.IsVip,
	})

	// 处理标签更新 (多对多)
	if len(req.Tags) > 0 {
		DB.Model(&blog).Association("Tags").Replace(req.Tags)
	}

	c.JSON(http.StatusOK, blog)
}

func deleteBlog(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var currentUser User
	DB.First(&currentUser, userID)

	var blog Blog
	if err := DB.First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "文章未找到"})
		return
	}

	if blog.AuthorID != userID && currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权删除他人文章"})
		return
	}

	DB.Delete(&blog)
	c.JSON(http.StatusOK, utils.H{"message": "删除成功"})
}

// 评论功能
func addComment(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var comment Comment
	c.BindAndValidate(&comment)
	comment.UserID = userID
	DB.Create(&comment)

	// 发放评论奖励
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
	all := c.Query("all") == "true"

	var currentUser User
	DB.First(&currentUser, userID)

	blogs := []Blog{}
	query := DB.Preload("Category").Preload("Tags").Preload("OriginalBlog").Preload("OriginalBlog.Author").Order("created_at desc")

	if all && currentUser.Role == "admin" {
		query.Find(&blogs)
	} else {
		query.Where("author_id = ?", userID).Find(&blogs)
	}
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
	logs := []RewardLog{}
	DB.Where("user_id = ?", userID).Order("created_at desc").Find(&logs)
	c.JSON(http.StatusOK, logs)
}

func shareBlog(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	blogID := c.Param("id")

	var blog Blog
	if err := DB.First(&blog, blogID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "文章未找到"})
		return
	}

	// 增加分享计数并同步更新
	DB.Model(&blog).Update("share_count", gorm.Expr("share_count + 1"))

	// 发放分享奖励 (复用 FORWARD 奖励配置)
	go DistributeRewards(userID, "FORWARD")

	c.JSON(http.StatusOK, utils.H{
		"message":     "分享成功，奖励已发放",
		"share_count": blog.ShareCount + 1,
	})
}

// 排行榜逻辑 (排除 admin 用户)
func getLeaderboard(ctx context.Context, c *app.RequestContext) {
	users := []User{}
	DB.Where("role != ?", "admin").Order("blog_balance desc").Limit(100).Find(&users)
	c.JSON(http.StatusOK, users)
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

	users := []User{}
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

func adminToggleVip(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var admin User
	DB.First(&admin, userID)
	if admin.Role != "admin" {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权访问此页面"})
		return
	}

	targetID := c.Param("id")
	var target User
	if err := DB.First(&target, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "用户未找到"})
		return
	}

	target.IsVip = !target.IsVip
	DB.Save(&target)

	c.JSON(http.StatusOK, utils.H{
		"message": "会员状态已更新",
		"is_vip":  target.IsVip,
	})
}

func applyVip(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var user User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "用户未找到"})
		return
	}

	if user.IsVip {
		c.JSON(http.StatusBadRequest, utils.H{"error": "您已经是会员"})
		return
	}

	var existing VipApplication
	if err := DB.Where("user_id = ? AND status = ?", userID, "pending").First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "您已有待处理的会员申请"})
		return
	}

	if user.BlogBalance < 1000 {
		c.JSON(http.StatusBadRequest, utils.H{"error": "代币余额不足，需要1000 BLOG"})
		return
	}

	var admin User
	DB.Where("role = ?", "admin").First(&admin)

	DB.Model(&user).Update("blog_balance", gorm.Expr("blog_balance - ?", 1000))
	if admin.ID != 0 {
		DB.Model(&admin).Update("blog_balance", gorm.Expr("blog_balance + ?", 1000))
	}

	app := VipApplication{UserID: userID, Amount: 1000}
	DB.Create(&app)

	c.JSON(http.StatusOK, utils.H{
		"message":        "会员申请已提交，请等待管理员审核",
		"application_id": app.ID,
	})
}

func getVipStatus(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var user User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "用户未找到"})
		return
	}

	var pendingApp VipApplication
	DB.Where("user_id = ? AND status = ?", userID, "pending").First(&pendingApp)

	c.JSON(http.StatusOK, utils.H{
		"is_vip":       user.IsVip,
		"has_pending":  pendingApp.ID != 0,
		"blog_balance": user.BlogBalance,
	})
}

func buyTokens(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var req struct {
		AmountUSD     float64 `json:"amount_usd"`
		PaymentMethod string  `json:"payment_method"`
	}
	if err := c.BindAndValidate(&req); err != nil || req.AmountUSD <= 0 {
		c.JSON(http.StatusBadRequest, utils.H{"error": "无效金额"})
		return
	}

	// 汇率: 1 USD = 10 BLOG
	blogAmount := req.AmountUSD * 10

	// 模拟支付成功 (实际生产环境应通过 Webhook 回调确认支付后发放)
	DB.Model(&User{}).Where("id = ?", userID).Update("blog_balance", gorm.Expr("blog_balance + ?", blogAmount))

	c.JSON(http.StatusOK, utils.H{
		"message":     "购买成功",
		"blog_amount": blogAmount,
	})
}

func adminGetVipApplications(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var admin User
	DB.First(&admin, userID)
	if admin.Role != "admin" {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权访问此页面"})
		return
	}

	applications := []VipApplication{}
	DB.Preload("User").Where("status = ?", "pending").Order("created_at desc").Find(&applications)
	c.JSON(http.StatusOK, applications)
}

func adminApproveVip(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var admin User
	DB.First(&admin, userID)
	if admin.Role != "admin" {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权访问此页面"})
		return
	}

	appID := c.Param("id")
	var app VipApplication
	if err := DB.First(&app, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "申请未找到"})
		return
	}

	if app.Status != "pending" {
		c.JSON(http.StatusBadRequest, utils.H{"error": "申请已处理"})
		return
	}

	app.Status = "approved"
	DB.Save(&app)

	var user User
	DB.First(&user, app.UserID)
	user.IsVip = true
	DB.Save(&user)

	c.JSON(http.StatusOK, utils.H{"message": "会员申请已批准"})
}

func adminRejectVip(ctx context.Context, c *app.RequestContext) {
	userID := c.GetUint("user_id")
	var admin User
	DB.First(&admin, userID)
	if admin.Role != "admin" {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权访问此页面"})
		return
	}

	appID := c.Param("id")
	var app VipApplication
	if err := DB.First(&app, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "申请未找到"})
		return
	}

	if app.Status != "pending" {
		c.JSON(http.StatusBadRequest, utils.H{"error": "申请已处理"})
		return
	}

	app.Status = "rejected"
	DB.Save(&app)

	var user User
	DB.First(&user, app.UserID)
	DB.Model(&user).Update("blog_balance", gorm.Expr("blog_balance + ?", app.Amount))

	var adminUser User
	DB.Where("role = ?", "admin").First(&adminUser)
	if adminUser.ID != 0 {
		DB.Model(&adminUser).Update("blog_balance", gorm.Expr("blog_balance - ?", app.Amount))
	}

	c.JSON(http.StatusOK, utils.H{"message": "会员申请已拒绝，代币已退还"})
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
