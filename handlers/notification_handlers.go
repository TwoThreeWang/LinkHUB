package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateNotification 创建新的通知
func CreateNotification(userID uint, content string, fromID uint) error {
	notification := &models.Notification{
		UserID:   userID,
		Content:  content,
		FromID:   fromID,
		Status:   0, // 默认未读状态
	}

	return database.GetDB().Create(notification).Error
}

// GetUserNotifications 获取用户的所有通知
func GetUserNotifications(userID uint) (notifications []models.Notification, err error) {
	err = database.GetDB().
		Joins("LEFT JOIN users ON notifications.from_id = users.id").
		Select("notifications.*, CASE WHEN notifications.from_id = 0 THEN 'System' ELSE users.username END as from_name").
		Where("notifications.user_id = ?", userID).
		Order("notifications.created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// GetUnreadCount 获取用户未读通知数量
func GetUnreadCount(userID uint) int64 {
	var count int64
	database.GetDB().Model(&models.Notification{}).
		Where("user_id = ? AND status = 0", userID).
		Count(&count)
	return count
}

// DeleteNotification 删除通知
func DeleteNotification(c *gin.Context) {
	// 获取通知ID
	notificationID := c.Param("id")
	if notificationID == "" {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "通知ID不能为空",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取当前用户
	user := GetCurrentUser(c)
	if user == nil {
		c.HTML(http.StatusUnauthorized, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "用户未登录",
			"redirect_text": "返回",
		}))
		return
	}

	// 删除通知
	// 将通知ID转换为uint类型
    nID, err := strconv.ParseUint(notificationID, 10, 64)
    if err != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "系统错误："+err.Error(),
			"redirect_text": "返回",
		}))
        return
    }

    // 删除通知，确保只能删除属于自己的通知
    result := database.GetDB().Where("id = ? AND user_id = ?", nID, user.ID).Delete(&models.Notification{})
    if result.Error != nil {
        c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "系统错误："+result.Error.Error(),
			"redirect_text": "返回",
		}))
		return
    }

	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	c.Redirect(302, refer)
}

// ReadNotification 标记通知为已读
func ReadNotification(c *gin.Context) {
	// 获取通知ID
	notificationID := c.Param("id")
	if notificationID == "" {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "通知ID不能为空",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取当前用户
	user := GetCurrentUser(c)
	if user == nil {
		c.HTML(http.StatusUnauthorized, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "用户未登录",
			"redirect_text": "返回",
		}))
		return
	}

	// 标记通知为已读
	err:=database.GetDB().Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, user.ID).
		Update("status", 1).Error
	if err != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "标记通知已读失败",
			"redirect_text": "返回",
		}))
		return
	}

	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	c.Redirect(302, refer)
}