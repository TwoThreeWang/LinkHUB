package handlers

import (
	"net/http"
	"strconv"
	"time"

	"LinkHUB/database"
	"LinkHUB/models"

	"github.com/gin-gonic/gin"
)

// CreateComment 创建评论
func CreateComment(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	user := userInterface.(models.User)

	// 获取表单数据
	linkID, err := strconv.Atoi(c.PostForm("link_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的链接ID"})
		return
	}
	content := c.PostForm("content")
	parentIDStr := c.PostForm("parent_id")

	// 验证评论内容
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}

	// 创建评论
	comment := models.Comment{
		LinkID:    uint(linkID),
		UserID:    user.ID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果有父评论ID，验证并设置父评论
	if parentIDStr != "" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的父评论ID"})
			return
		}

		// 验证父评论是否存在
		var parentComment models.Comment
		if err := database.GetDB().First(&parentComment, parentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父评论不存在"})
			return
		}

		// 确保父评论属于同一个链接
		if parentComment.LinkID != uint(linkID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "父评论必须属于同一个链接"})
			return
		}

		// 将uint类型转换为*uint类型
		parentIDUint := uint(parentID)
		comment.ParentID = &parentIDUint
	}

	// 保存评论
	if err := database.GetDB().Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "评论创建成功", "comment": comment})
}