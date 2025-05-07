package models

import (
	"gorm.io/gorm"
	"time"
)

// Ads 广告模型
type Ads struct {
	gorm.Model
	Name    string    `gorm:"type:varchar(500)" json:"name"`            // 广告标题
	Url     string    `gorm:"size:500;not null" json:"url"`             // 广告链接
	Img     string    `gorm:"type:varchar(500)" json:"img"`             // 广告图片
	AdType  string    `gorm:"type:varchar(10);not null" json:"ad_type"` // 广告类型
	Email   string    `gorm:"type:varchar(100);not null" json:"email"`  // 联系邮箱
	EndDate time.Time `gorm:"not null" json:"end_date"`                 // 到期日期
}
