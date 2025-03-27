package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateArticleComment 创建文章评论
func CreateArticleComment(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "用户未登录",
			"redirect_text": "去登录",
			"refer":         "/auth/login",
		}))
		return
	}

	// 获取表单数据
	articleID, err := strconv.Atoi(c.PostForm("article_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "无效的文章ID",
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
	comment := models.ArticleComment{
		ArticleID: uint(articleID),
		UserID:    userInfo.ID,
		Content:   content,
	}

	// 如果有父评论ID，验证并设置父评论
	var parentComment models.ArticleComment
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
		if err := database.GetDB().First(&parentComment, parentID).Error; err != nil {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "父评论不存在",
				"redirect_text": "返回",
			}))
			return
		}

		// 确保父评论属于同一个文章
		if parentComment.ArticleID != uint(articleID) {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "父评论必须属于同一个文章",
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
	// 发送消息
	go func() {
		// 根据ArticleID查询文章标题
		var article models.Article
		if err := database.GetDB().First(&article, uint(articleID)).Error; err != nil {
			return
		}
		// 如果是回复评论，发送消息给父评论的作者
		if parentComment.ID != 0 && parentComment.UserID!= userInfo.ID{
			content := fmt.Sprintf("<a href='/articles/%d'>您在文章《%s》上的评论有新回复了，点击查看</a>", article.ID, article.Title)
			_ = CreateNotification(parentComment.UserID, content, 0)
		}
		// 如果评论自己的文章， 就不用通知了;如果父评论和文章作者是同一个人，只保留上边的通知就行了
		if article.UserID != userInfo.ID && parentComment.UserID != article.UserID {
			content := fmt.Sprintf("<a href='/articles/%d'>您的文章《%s》有新评论了，点击查看</a>", article.ID, article.Title)
			_ = CreateNotification(article.UserID, content, 0)
		}

	}()
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	// 重定向到指定页面
	c.Redirect(http.StatusFound, refer)
}
