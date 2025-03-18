package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Username string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"username"`
	Email    string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"email"`
	Password string    `gorm:"type:varchar(100);not null" json:"-"`
	Avatar   string    `gorm:"type:varchar(500);" json:"avatar"`
	Bio      string    `gorm:"type:varchar(150);" json:"bio"`
	Role     string    `gorm:"column:role;type:varchar(20)" json:"role"`
	Links    []Link    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"links,omitempty"`
	Votes    []Vote    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"votes,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	Articles []Article `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"articles,omitempty"`
}

// BeforeSave 保存前的钩子函数，用于加密密码
func (u *User) BeforeSave(tx *gorm.DB) error {
	// 如果密码已经被加密，则不再加密
	if len(u.Password) > 0 && len(u.Password) < 60 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword 检查密码是否正确
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
