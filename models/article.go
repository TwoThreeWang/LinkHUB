package models

import (
	"gorm.io/gorm"
)

// Article 文章模型
type Article struct {
	gorm.Model
	Title      string           `gorm:"size:200;not null" json:"title"`
	Content    string           `gorm:"type:text;not null" json:"content"`
	UserID     uint             `gorm:"not null" json:"user_id"`
	User       User             `json:"user,omitempty"`
	ViewCount  int              `gorm:"default:0" json:"view_count"`
	CategoryID uint             `gorm:"" json:"category_id"`
	Category   Category         `json:"category,omitempty"`
	Comments   []ArticleComment `gorm:"constraint:OnDelete:CASCADE;" json:"comments,omitempty"`
}

// IncreaseViewCount 增加浏览量
func (a *Article) IncreaseViewCount() {
	a.ViewCount++
}
