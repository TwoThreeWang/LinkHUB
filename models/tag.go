package models

import (
	"gorm.io/gorm"
)

// Tag 标签模型
type Tag struct {
	gorm.Model
	Name  string `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Count int    `gorm:"default:0" json:"count"`
	Links []Link `gorm:"many2many:link_tags;" json:"links,omitempty"`
}

// IncreaseCount 增加标签使用计数
func (t *Tag) IncreaseCount() {
	t.Count++
}

// DecreaseCount 减少标签使用计数
func (t *Tag) DecreaseCount() {
	t.Count--
}
