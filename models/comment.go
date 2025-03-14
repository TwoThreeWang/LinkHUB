package models

import (
	"gorm.io/gorm"
)

// Comment 评论模型
type Comment struct {
	gorm.Model
	Content  string    `gorm:"size:1000;not null" json:"content"`
	UserID   uint      `gorm:"not null" json:"user_id"`
	LinkID   uint      `gorm:"not null" json:"link_id"`
	ParentID *uint     `json:"parent_id"`
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Link     Link      `gorm:"foreignKey:LinkID" json:"link,omitempty"`
	Replies  []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// ArticleComment 文章评论模型
type ArticleComment struct {
	gorm.Model
	Content   string           `gorm:"size:1000;not null" json:"content"`
	UserID    uint             `gorm:"not null" json:"user_id"`
	ArticleID uint             `gorm:"not null" json:"article_id"`
	ParentID  *uint            `json:"parent_id"`
	User      User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Article   Article          `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	Replies   []ArticleComment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}
