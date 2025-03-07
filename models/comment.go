package models

import (
	"time"
)

// Comment 评论模型
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   string    `gorm:"size:1000;not null" json:"content"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	LinkID    uint      `gorm:"not null" json:"link_id"`
	ParentID  *uint     `json:"parent_id"` // 父评论ID，用于回复功能
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user,omitempty"`
	Link      Link      `json:"link,omitempty"`
	Replies   []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName 设置表名
func (Comment) TableName() string {
	return "comments"
}