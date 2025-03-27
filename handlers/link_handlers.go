package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"LinkHUB/database"
	"LinkHUB/models"
	"LinkHUB/utils"

	"github.com/gin-gonic/gin"
)

// Home 首页处理函数
func Home(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// 获取排序参数
	sort := c.DefaultQuery("sort", "new")

	// 获取链接列表
	var links []models.Link
	var total int64

	// 构建查询
	query := database.GetDB().Model(&models.Link{})

	// 计算总数
	query.Count(&total)

	// 获取分页数据
	pageSize := 10
	offset := (page - 1) * pageSize

	// 根据排序参数设置排序方式
	// 首先按照置顶状态排序
	query = query.Order("is_pinned DESC")

	switch sort {
	case "top":
		// 使用vote_count和click_count的和进行排序
		query = query.Order("(vote_count + click_count) DESC")
	case "new":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// 执行查询
	query.Limit(pageSize).
		Offset(offset).
		Preload("User").
		Preload("Tags").
		Find(&links)

	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize

	// 获取热门标签
	var popularTags []models.Tag
	database.GetDB().Order("count DESC").Limit(10).Find(&popularTags)

	// 渲染模板
	c.HTML(http.StatusOK, "home", OutputCommonSession(c, gin.H{
		"title":       "发现精彩链接",
		"links":       links,
		"page":        page,
		"totalPages":  totalPages,
		"sort":        sort,
		"popularTags": popularTags,
	}))
}

// ShowNewLink 显示创建链接页面
func ShowNewLink(c *gin.Context) {
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
	// 查询所有标签
	var tags []models.Tag
	database.GetDB().Find(&tags)

	// 渲染模板
	c.HTML(http.StatusOK, "new_link", OutputCommonSession(c, gin.H{
		"title": "分享新链接",
		"tags":  tags,
	}))
}

// CreateLink 创建链接处理函数
func CreateLink(c *gin.Context) {
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
	title := c.PostForm("title")
	url := c.PostForm("url")
	description := c.PostForm("description")
	checkTags := c.PostFormArray("tags[]")
	if len(checkTags) > 5 {
		checkTags = checkTags[:5]
	}

	// 查询所有标签
	var tags []models.Tag
	database.GetDB().Find(&tags)

	// 验证表单数据
	if title == "" || url == "" {
		c.HTML(http.StatusBadRequest, "new_link", OutputCommonSession(c, gin.H{
			"title":       "分享新链接",
			"error":       "标题和URL是必填的",
			"link_title":  title,
			"url":         url,
			"description": description,
			"tags":        tags,
			"checkTags":   checkTags,
		}))
		return
	}

	// 创建链接
	link := models.Link{
		Title:       title,
		URL:         url,
		Description: description,
		UserID:      userInfo.ID,
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 保存链接
	if err := tx.Create(&link).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
			"title":       "分享新链接",
			"error":       "创建链接失败: " + err.Error(),
			"link_title":  title,
			"url":         url,
			"description": description,
			"tags":        tags,
			"checkTags":   checkTags,
		}))
		return
	}

	// 处理每个标签
	for _, name := range checkTags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// 查找或创建标签
		var tag models.Tag
		result := tx.Where("name = ?", name).FirstOrCreate(&tag, models.Tag{
			Name: name,
		})

		if result.Error != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
				"title":       "分享新链接",
				"error":       "处理标签失败: " + result.Error.Error(),
				"link_title":  title,
				"url":         url,
				"description": description,
				"tags":        tags,
				"checkTags":   checkTags,
			}))
			return
		}

		// 增加标签计数
		if result.RowsAffected == 0 {
			tx.Model(&tag).Update("count", tag.Count+1)
		} else {
			tag.Count = 1
			tx.Save(&tag)
		}

		// 关联标签和链接
		if err := tx.Model(&link).Association("Tags").Append(&tag); err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
				"title":       "分享新链接",
				"error":       "关联标签失败: " + err.Error(),
				"link_title":  title,
				"url":         url,
				"description": description,
				"tags":        tags,
				"checkTags":   checkTags,
			}))
			return
		}
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
			"title":       "分享新链接",
			"error":       "保存链接失败: " + err.Error(),
			"link_title":  title,
			"url":         url,
			"description": description,
			"tags":        tags,
			"checkTags":   checkTags,
		}))
		return
	}

	// 重定向到链接详情页
	c.Redirect(http.StatusFound, "/links/"+strconv.Itoa(int(link.ID)))
}

// ShowLink 显示链接详情页面
func ShowLink(c *gin.Context) {
	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().Preload("User").Preload("Tags").First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "链接不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取评论
	var comments []models.Comment
	database.GetDB().Where("link_id = ? AND parent_id IS NULL", link.ID).
		Preload("User").
		Preload("Replies").
		Preload("Replies.User").
		Order("created_at ASC").
		Find(&comments)

	// 检查当前用户是否已投票
	userInfo := GetCurrentUser(c)
	voted := false
	if userInfo != nil {
		var count int64
		database.GetDB().Model(&models.Vote{}).Where("user_id = ? AND link_id = ?", userInfo.ID, link.ID).Count(&count)
		voted = count > 0
	}
	// 相关链接
	var relatedLinks []models.Link
	if len(link.Tags) > 0 {
		// 获取当前链接的标签IDs
		var tagIDs []uint
		for _, tag := range link.Tags {
			tagIDs = append(tagIDs, tag.ID)
		}

		// 查询具有相同标签的其他链接（不包括当前链接）
		database.GetDB().Distinct().
			Joins("JOIN link_tags ON link_tags.link_id = links.id").
			Where("link_tags.tag_id IN (?) AND links.id != ?", tagIDs, link.ID).
			Preload("User").
			Preload("Tags").
			Limit(6).
			Find(&relatedLinks)
	}

	// 渲染模板
	c.HTML(http.StatusOK, "link_detail", OutputCommonSession(c, gin.H{
		"title":        link.Title,
		"link":         link,
		"comments":     comments,
		"voted":        voted,
		"relatedLinks": relatedLinks,
	}))
}

// ShowUpdateLink 显示修改链接页面
func ShowUpdateLink(c *gin.Context) {
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
	// 获取链接ID
	id := c.Param("id")
	// 查询链接
	var link models.Link
	result := database.GetDB().Preload("Tags").First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "链接不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}

	// 验证用户权限
	if link.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "您没有权限编辑此链接",
			"redirect_text": "返回",
		}))
		return
	}
	// 查询所有标签
	var tags []models.Tag
	database.GetDB().Find(&tags)

	var checkTags []string
	for _, tag := range link.Tags {
		checkTags = append(checkTags, tag.Name)
	}

	// 渲染模板
	c.HTML(http.StatusOK, "new_link", OutputCommonSession(c, gin.H{
		"title":       "编辑链接",
		"id":          id,
		"link_title":  link.Title,
		"url":         link.URL,
		"description": link.Description,
		"tags":        tags,
		"checkTags":   checkTags,
	}))
}

// TogglePinLink 切换链接置顶状态
func TogglePinLink(c *gin.Context) {
	// 获取链接ID
	id := c.Param("id")
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, OutputApi(403, "用户未登录"))
		return
	}

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, OutputApi(404, "链接不存在或已被删除"))
		return
	}

	// 验证用户权限
	if link.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.JSON(http.StatusForbidden, OutputApi(403, "您没有权限操作此链接"))
		return
	}

	// 切换置顶状态
	link.IsPinned = !link.IsPinned
	if err := database.GetDB().Save(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, OutputApi(400, "操作失败"))
		return
	}

	c.JSON(http.StatusOK, OutputApi(200, "操作成功"))
}

// UpdateLink 更新链接处理函数
func UpdateLink(c *gin.Context) {
	// 获取链接ID
	id := c.Param("id")
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "请先登录",
			"redirect_text": "返回",
			"refer":         "/links/" + id,
		}))
		return
	}

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "链接不存在或已被删除",
			"redirect_text": "返回",
			"refer":         "/links/" + id,
		}))
		return
	}

	// 验证用户权限
	if link.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "您没有权限编辑此链接",
			"redirect_text": "返回",
			"refer":         "/links/" + id,
		}))
		return
	}

	// 获取表单数据
	title := c.PostForm("title")
	url := c.PostForm("url")
	description := c.PostForm("description")
	checkTags := c.PostFormArray("tags[]")
	if len(checkTags) > 5 {
		checkTags = checkTags[:5]
	}

	// 查询所有标签
	var tags []models.Tag
	database.GetDB().Find(&tags)

	// 验证表单数据
	if title == "" || url == "" {
		c.HTML(http.StatusBadRequest, "new_link", OutputCommonSession(c, gin.H{
			"title":       "编辑链接",
			"error":       "标题和URL是必填的",
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"tags":        tags,
			"checkTags":   checkTags,
		}))
		return
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 更新链接信息
	link.Title = title
	link.URL = url
	link.Description = description
	link.UpdatedAt = time.Now()

	// 保存链接
	if err := tx.Save(&link).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
			"title":       "编辑链接",
			"error":       "更新链接失败: " + err.Error(),
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"checkTags":   checkTags,
			"tags":        tags,
		}))
		return
	}

	// 清除现有标签关联
	if err := tx.Model(&link).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
			"title":       "编辑链接",
			"error":       "清除标签关联失败: " + err.Error(),
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"checkTags":   checkTags,
			"tags":        tags,
		}))
		return
	}

	// 处理每个标签
	for _, name := range checkTags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// 查找或创建标签
		var tag models.Tag
		result := tx.Where("name = ?", name).FirstOrCreate(&tag, models.Tag{
			Name: name,
		})

		if result.Error != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
				"title":       "编辑链接",
				"error":       "处理标签失败: " + result.Error.Error(),
				"link":        link,
				"link_title":  title,
				"url":         url,
				"description": description,
				"checkTags":   checkTags,
				"tags":        tags,
			}))
			return
		}

		// 增加标签计数
		if result.RowsAffected == 0 {
			tx.Model(&tag).Update("count", tag.Count+1)
		} else {
			tag.Count = 1
			tx.Save(&tag)
		}

		// 关联标签和链接
		if err := tx.Model(&link).Association("Tags").Append(&tag); err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
				"title":       "编辑链接",
				"error":       "关联标签失败: " + err.Error(),
				"link":        link,
				"link_title":  title,
				"url":         url,
				"description": description,
				"checkTags":   checkTags,
				"tags":        tags,
			}))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "new_link", OutputCommonSession(c, gin.H{
			"title":       "编辑链接",
			"error":       "保存链接失败: " + err.Error(),
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"checkTags":   checkTags,
			"tags":        tags,
		}))
		return
	}

	// 重定向到链接详情页
	c.Redirect(http.StatusFound, "/links/"+strconv.Itoa(int(link.ID)))
}

// DeleteLink 删除链接处理函数
func DeleteLink(c *gin.Context) {
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

	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "链接不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}

	// 验证用户权限
	if link.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "您没有权限删除此链接",
			"redirect_text": "返回",
		}))
		return
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	tx.Model(&link).Association("Tags").Clear()

	// 清除标签关联
	if err := tx.Model(&link).Association("Tags").Unscoped().Delete(); err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "清除标签关联失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	//tx.Model(&link).Association("Votes").Clear()

	// 清除投票关联
	if err := tx.Model(&link).Association("Votes").Unscoped().Delete(); err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "清除投票关联失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 删除链接
	if err := tx.Unscoped().Delete(&link).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "删除链接失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "删除链接失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 返回成功响应
	c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
		"title":         "Success",
		"message":       "删除链接成功",
		"redirect_text": "返回首页",
		"refer":         "/",
	}))
}

// VoteLink 投票链接处理函数
func VoteLink(c *gin.Context) {
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

	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "链接不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}

	// 检查用户是否已投票
	var count int64
	database.GetDB().Model(&models.Vote{}).Where("user_id = ? AND link_id = ?", userInfo.ID, link.ID).Count(&count)
	if count > 0 {
		c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
			"title":         "Warning",
			"message":       "您已经投过票了",
			"redirect_text": "返回",
		}))
		return
	}

	// 创建投票记录
	vote := models.Vote{
		UserID: userInfo.ID,
		LinkID: link.ID,
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 保存投票
	if err := tx.Create(&vote).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "投票失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 更新链接投票计数
	if err := tx.Model(&link).Update("vote_count", link.VoteCount+1).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "更新投票计数失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "投票失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}
	// 发送消息
	if link.UserID!= userInfo.ID {
		go func() {
			content := fmt.Sprintf("您的链接《<a href='/links/%d'>%s</a>》被用户 <a href='/user/profile/%d'>%s</a> 投票了", link.ID, link.Title, userInfo.ID, userInfo.Username)
			_ = CreateNotification(link.UserID, content, 0)
		}()
	}
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	c.Redirect(302, refer)
}

// UnVoteLink 取消投票处理函数
func UnVoteLink(c *gin.Context) {
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

	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "链接不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}

	// 检查用户是否已投票
	var vote models.Vote
	result = database.GetDB().Where("user_id = ? AND link_id = ?", userInfo.ID, link.ID).First(&vote)
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "您还没有投过票",
			"redirect_text": "返回",
		}))
		return
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 删除投票记录
	if err := tx.Delete(&vote).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "取消投票失败",
			"redirect_text": "返回",
		}))
		return
	}

	// 更新链接投票计数
	if err := tx.Model(&link).Update("vote_count", link.VoteCount-1).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "更新投票计数失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "取消投票失败: " + err.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 返回成功响应
	c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
		"title":         "Success",
		"message":       "已成功取消投票",
		"redirect_text": "返回",
	}))
}

// ClickLink 点击链接处理函数
func ClickLink(c *gin.Context) {
	// 获取链接ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusOK, OutputApi(400, "参数错误"))
		return
	}

	// 获取客户端IP
	clientIP := c.ClientIP()
	cacheKey := fmt.Sprintf("%s%s", clientIP, id)

	// 检查是否是新点击（同一IP 24小时内对同一链接只记录一次）
	_, isClicked := utils.GlobalCache.Get(cacheKey)
	if isClicked {
		c.JSON(http.StatusOK, OutputApi(400, "重复点击"))
		return
	}

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.JSON(http.StatusOK, OutputApi(400, "参数错误，未查询到链接ID"))
		return
	}

	// 更新链接点击计数
	if err := database.GetDB().Model(&link).Update("click_count", link.ClickCount+1).Error; err != nil {
		c.JSON(http.StatusOK, OutputApi(400, "更新链接点击计数失败: "+err.Error()))
		return
	} else {
		// 设置缓存
		utils.GlobalCache.Set(cacheKey, 1, time.Hour*12)
	}

	// 重定向到链接URL
	c.JSON(http.StatusOK, OutputApi(200, "Success"))
}

// SearchLinks 搜索处理函数
func SearchLinks(c *gin.Context) {
	// 获取查询参数
	query := c.Query("q")
	if query == "" {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "搜索词不能为空",
			"redirect_text": "返回",
		}))
		return
	}
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	// 获取分页数据
	pageSize := 10
	offset := (page - 1) * pageSize
	// 搜索链接
	var links []models.Link
	var total int64
	// 构建查询条件
	queryDB := database.GetDB().Where(
		"title ILIKE ? OR description ILIKE ? OR url ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%",
	)
	// 先计算总记录数
	queryDB.Model(&models.Link{}).Count(&total)
	// 执行分页查询
	queryDB.Preload("Tags").Preload("User").Limit(pageSize).Offset(offset).Find(&links)
	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize
	// 渲染模板
	c.HTML(http.StatusOK, "search", OutputCommonSession(c, gin.H{
		"title":      query,
		"query":      query,
		"links":      links,
		"page":       page,
		"totalPages": totalPages,
		"total":      total,
	}))
}
