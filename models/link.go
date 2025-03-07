package models

import (
	"time"
)

// Link 链接模型
type Link struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	URL         string    `gorm:"size:500;not null" json:"url"`
	Description string    `gorm:"size:1000" json:"description"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `json:"user,omitempty"`
	VoteCount   int       `gorm:"default:0" json:"vote_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []Tag     `gorm:"many2many:link_tags;" json:"tags,omitempty"`
	Votes       []Vote    `json:"votes,omitempty"`
	Comments    []Comment `json:"comments,omitempty"`
}

// IncreaseVoteCount 增加投票数
func (l *Link) IncreaseVoteCount() {
	l.VoteCount++
}

// DecreaseVoteCount 减少投票数
func (l *Link) DecreaseVoteCount() {
	l.VoteCount--
}
