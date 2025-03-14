package models

import (
	"gorm.io/gorm"
)

// Category 分类模型
type Category struct {
	gorm.Model
	Name     string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Count    int       `gorm:"default:0" json:"count"`
	Articles []Article `json:"articles,omitempty"`
}

// IncreaseCount 增加分类使用计数
func (c *Category) IncreaseCount() {
	c.Count++
}

// DecreaseCount 减少分类使用计数
func (c *Category) DecreaseCount() {
	c.Count--
}
