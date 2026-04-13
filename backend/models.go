package main

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Username      string         `gorm:"unique;not null" json:"username"`
	PasswordHash  string         `gorm:"not null" json:"-"`
	Email         string         `gorm:"unique;not null" json:"email"`
	Avatar        string         `json:"avatar"`
	Role          string         `gorm:"default:'user'" json:"role"`
	IsVip         bool           `gorm:"default:false" json:"is_vip"`
	WalletAddress string         `gorm:"unique" json:"wallet_address"`
	ReferralCode  string         `json:"referral_code"`
	ReferrerID    *uint          `json:"referrer_id"`
	BlogBalance   float64        `gorm:"default:0" json:"blog_balance"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type RewardLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	Type      string    `json:"type"` // Register/Post/Comment/Forward/Referral
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"unique;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Blog struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Title        string         `gorm:"not null" json:"title"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	ImageURL     string         `json:"image_url"`
	VideoURL     string         `json:"video_url"`
	AuthorID     uint           `json:"author_id"`
	Author       User           `gorm:"foreignKey:AuthorID" json:"author"`
	CategoryID   *uint          `json:"category_id"`
	Category     Category       `gorm:"foreignKey:CategoryID" json:"category"`
	Tags         []Tag          `gorm:"many2many:blog_tags;" json:"tags"` // 多对多关联
	Status       string         `gorm:"default:'published'" json:"status"`
	IsVip        bool           `gorm:"default:false" json:"is_vip"`
	ViewCount    int            `gorm:"default:0" json:"view_count"`
	ShareCount   int            `gorm:"default:0" json:"share_count"`
	IsForward    bool           `gorm:"default:false" json:"is_forward"`
	OriginalID   *uint          `json:"original_id"`
	OriginalBlog *Blog          `gorm:"foreignKey:OriginalID" json:"original_blog"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BlogID    uint           `json:"blog_id"`
	UserID    uint           `json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	ParentID  *uint          `json:"parent_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Media struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FileName   string    `json:"file_name"`
	FileURL    string    `json:"file_url"`
	FileType   string    `json:"file_type"`
	UploaderID uint      `json:"uploader_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type VipApplication struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	Status    string         `gorm:"default:'pending'" json:"status"` // pending, approved, rejected
	Amount    float64        `gorm:"default:1000" json:"amount"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
