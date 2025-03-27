package models

import "time"

// Notification 通知消息模型
type Notification struct {
    ID        uint      `json:"id" gorm:"primarykey"`
    UserID    uint      `json:"user_id" gorm:"index"` // 接收通知的用户ID
    Content   string    `json:"content"`              // 通知内容
    FromID    uint      `json:"from_id"`             // 发送通知的用户ID
    FromName  string    `json:"from_name"`   // 发送者用户名
    Status    int       `json:"status"`              // 通知状态：0-未读，1-已读
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}