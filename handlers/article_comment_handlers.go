package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateArticleComment 创建文章评论
func CreateArticleComment(c *gin.Context) {
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"title":         "Error",
			"message":       "请先登录",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 获取表单数据
	articleID, err := strconv.Atoi(c.PostForm("article_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "Error",
			"message":       "无效的文章ID",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}
	content := c.PostForm("content")
	parentIDStr := c.PostForm("parent_id")

	// 验证评论内容
	if content == "" {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "Error",
			"message":       "评论内容不能为空",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 创建评论
	comment := models.ArticleComment{
		ArticleID: uint(articleID),
		UserID:    userInfo.ID,
		Content:   content,
	}

	// 如果有父评论ID，验证并设置父评论
	if parentIDStr != "" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "result", gin.H{
				"userInfo":      userInfo,
				"title":         "Error",
				"message":       "无效的父评论ID",
				"redirect_text": "返回",
				"redirect_url":  refer,
			})
			return
		}

		// 验证父评论是否存在
		var parentComment models.ArticleComment
		if err := database.GetDB().First(&parentComment, parentID).Error; err != nil {
			c.HTML(http.StatusBadRequest, "result", gin.H{
				"userInfo":      userInfo,
				"title":         "Error",
				"message":       "父评论不存在",
				"redirect_text": "返回",
				"redirect_url":  refer,
			})
			return
		}

		// 确保父评论属于同一个文章
		if parentComment.ArticleID != uint(articleID) {
			c.HTML(http.StatusBadRequest, "result", gin.H{
				"userInfo":      userInfo,
				"title":         "Error",
				"message":       "父评论必须属于同一个文章",
				"redirect_text": "返回",
				"redirect_url":  refer,
			})
			return
		}

		// 正确设置父评论ID（指针类型）
		parentIDUint := uint(parentID)
		comment.ParentID = &parentIDUint
	}

	// 保存评论
	if err := database.GetDB().Create(&comment).Error; err != nil {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "Error",
			"message":       "创建评论失败: " + err.Error(),
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 重定向到指定页面
	c.Redirect(http.StatusFound, refer)
}
