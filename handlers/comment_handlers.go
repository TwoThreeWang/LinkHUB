package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateComment 创建评论
func CreateComment(c *gin.Context) {
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "请先登录",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取表单数据
	linkID, err := strconv.Atoi(c.PostForm("link_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "无效的链接ID",
			"redirect_text": "返回",
		}))
		return
	}
	content := c.PostForm("content")
	parentIDStr := c.PostForm("parent_id")

	// 验证评论内容
	if content == "" {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "评论内容不能为空",
			"redirect_text": "返回",
		}))
		return
	}

	// 创建评论
	comment := models.Comment{
		LinkID:  uint(linkID),
		UserID:  userInfo.ID,
		Content: content,
	}

	// 如果有父评论ID，验证并设置父评论
	if parentIDStr != "" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "无效的父评论ID",
				"redirect_text": "返回",
			}))
			return
		}

		// 验证父评论是否存在
		var parentComment models.Comment
		if err := database.GetDB().First(&parentComment, parentID).Error; err != nil {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "父评论不存在",
				"redirect_text": "返回",
			}))
			return
		}

		// 确保父评论属于同一个链接
		if parentComment.LinkID != uint(linkID) {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "父评论必须属于同一个链接",
				"redirect_text": "返回",
			}))
			return
		}

		// 正确设置父评论ID（指针类型）
		parentIDUint := uint(parentID)
		comment.ParentID = &parentIDUint
	}

	// 保存评论
	if err := database.GetDB().Create(&comment).Error; err != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "创建评论失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 重定向到指定页面
	c.Redirect(http.StatusFound, refer)
}
