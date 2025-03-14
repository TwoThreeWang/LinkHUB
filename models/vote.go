package models

import (
	"fmt"

	"gorm.io/gorm"
)

// Vote 投票模型
type Vote struct {
	gorm.Model
	UserID uint `gorm:"not null" json:"user_id"`
	LinkID uint `gorm:"not null" json:"link_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Link   Link `gorm:"foreignKey:LinkID" json:"link,omitempty"`
}

// BeforeCreate 创建前的钩子函数
func (v *Vote) BeforeCreate(tx *gorm.DB) error {
	// 检查是否已经投票
	var count int64
	err := tx.Model(&Vote{}).Where("user_id = ? AND link_id = ?", v.UserID, v.LinkID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("用户已经对该链接投过票")
	}
	return nil
}

// AfterCreate 创建后的钩子函数
func (v *Vote) AfterCreate(tx *gorm.DB) error {
	// 增加链接的投票数
	return tx.Model(&Link{}).Where("id = ?", v.LinkID).Update("vote_count", gorm.Expr("vote_count + ?", 1)).Error
}

// AfterDelete 删除后的钩子函数
func (v *Vote) AfterDelete(tx *gorm.DB) error {
	// 减少链接的投票数
	return tx.Model(&Link{}).Where("id = ?", v.LinkID).Update("vote_count", gorm.Expr("vote_count - ?", 1)).Error
}
