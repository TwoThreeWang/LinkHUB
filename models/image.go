package models

import (
	"gorm.io/gorm"
)

// Image 图片模型
type Image struct {
	gorm.Model
	UserID      uint   `gorm:"not null" json:"user_id"`
	StorageType string `gorm:"type:varchar(25)" json:"storage_type"` // 存储类型
	ImageName   string `gorm:"type:varchar(50)" json:"image_name"`   // 图片名称
	ImageURL    string `gorm:"type:varchar(255)" json:"image_url"`   // 图片URL
	DeleteHash  string `gorm:"type:varchar(50)" json:"delete_hash"`  // 删除哈希
	FileSize    int64  `gorm:"not null" json:"file_size"`
	User        User   `gorm:"foreignKey:UserID" json:"user"`
}
