package models

import (
	"time"
)

// Tag 标签模型
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Count     int       `gorm:"default:0" json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Links     []Link    `gorm:"many2many:link_tags;" json:"links,omitempty"`
}

// IncreaseCount 增加标签使用计数
func (t *Tag) IncreaseCount() {
	t.Count++
}

// DecreaseCount 减少标签使用计数
func (t *Tag) DecreaseCount() {
	t.Count--
}
