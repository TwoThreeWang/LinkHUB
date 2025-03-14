package models

import (
	"gorm.io/gorm"
)

// Link 链接模型
type Link struct {
	gorm.Model
	Title       string    `gorm:"size:200;not null" json:"title"`
	URL         string    `gorm:"size:500;not null" json:"url"`
	Description string    `gorm:"size:1000" json:"description"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `json:"user,omitempty"`
	VoteCount   int       `gorm:"default:0" json:"vote_count"`
	ClickCount  int       `gorm:"default:0" json:"click_count"`
	Tags        []Tag     `gorm:"many2many:link_tags;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`
	Votes       []Vote    `gorm:"constraint:OnDelete:CASCADE;" json:"votes,omitempty"`
	Comments    []Comment `gorm:"constraint:OnDelete:CASCADE;" json:"comments,omitempty"`
}

// IncreaseVoteCount 增加投票数
func (l *Link) IncreaseVoteCount() {
	l.VoteCount++
}

// DecreaseVoteCount 减少投票数
func (l *Link) DecreaseVoteCount() {
	l.VoteCount--
}
