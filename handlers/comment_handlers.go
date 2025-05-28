package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"LinkHUB/utils"
	"fmt"
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
	CfTurnstile := c.PostForm("cf_turnstile")

	// 验证 Turnstile 令牌
	if CfTurnstile != "" {
		remoteIP := c.ClientIP()
		_, err := utils.VerifyTurnstileToken(c, CfTurnstile, remoteIP)
		if err!= nil {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "验证 Turnstile 令牌失败：" + err.Error(),
				"redirect_text": "返回",
			}))
			return
		}
	}else{
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "验证 Turnstile 令牌失败：缺少验证参数",
			"redirect_text": "返回",
		}))
		return
	}

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
	var parentComment models.Comment
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
	// 发送消息
	go func() {
		// 根据linkID查询文章标题
		var link models.Link
		if err := database.GetDB().First(&link, uint(linkID)).Error; err != nil {
			return
		}
		// 如果是回复评论，发送消息给父评论的作者
		if parentComment.ID != 0 && parentComment.UserID!= userInfo.ID{
			content := fmt.Sprintf("<a href='/links/%d'>您在链接《%s》上的评论有新回复了，点击查看</a>", link.ID, link.Title)
			_ = CreateNotification(parentComment.UserID, content, 0)
		}
		// 如果评论自己的文章， 就不用通知了;如果父评论和文章作者是同一个人，只保留上边的通知就行了
		if link.UserID != userInfo.ID && parentComment.UserID != link.UserID{
			content := fmt.Sprintf("<a href='/links/%d'>您的链接《%s》有新评论了，点击查看</a>", link.ID, link.Title)
			_ = CreateNotification(link.UserID, content, 0)
		}
	}()
	// 重定向到指定页面
	c.Redirect(http.StatusFound, refer)
}
