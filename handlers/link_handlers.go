package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"LinkHUB/database"
	"LinkHUB/models"

	"github.com/gin-gonic/gin"
)

// Home 首页处理函数
func Home(c *gin.Context) {
	userinfo := GetCurrentUser(c)
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
	switch sort {
	case "top":
		query = query.Order("vote_count DESC")
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
	c.HTML(http.StatusOK, "home", gin.H{
		"title":       "LinkHUB - 发现精彩链接",
		"links":       links,
		"page":        page,
		"totalPages":  totalPages,
		"sort":        sort,
		"popularTags": popularTags,
		"userInfo":        userinfo,
	})
}

// ListLinks 链接列表处理函数
func ListLinks(c *gin.Context) {
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
	switch sort {
	case "top":
		query = query.Order("vote_count DESC")
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

	// 渲染模板
	c.HTML(http.StatusOK, "links", gin.H{
		"title":      "所有链接 - LinkHUB",
		"links":      links,
		"page":       page,
		"totalPages": totalPages,
		"sort":       sort,
	})
}

// ShowNewLink 显示创建链接页面
func ShowNewLink(c *gin.Context) {
	c.HTML(http.StatusOK, "new_link", gin.H{
		"title": "分享新链接 - LinkHUB",
	})
}

// CreateLink 创建链接处理函数
func CreateLink(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		c.Abort()
		return
	}
	user := userInterface.(*models.User)

	// 获取表单数据
	title := c.PostForm("title")
	url := c.PostForm("url")
	description := c.PostForm("description")
	tagsStr := c.PostForm("tags")

	// 验证表单数据
	if title == "" || url == "" {
		c.HTML(http.StatusBadRequest, "new_link", gin.H{
			"title":       "分享新链接 - LinkHUB",
			"error":       "标题和URL是必填的",
			"link_title":  title,
			"url":         url,
			"description": description,
			"tags":        tagsStr,
		})
		return
	}

	// 创建链接
	link := models.Link{
		Title:       title,
		URL:         url,
		Description: description,
		UserID:      user.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 保存链接
	if err := tx.Create(&link).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "new_link", gin.H{
			"title":       "分享新链接 - LinkHUB",
			"error":       "创建链接失败: " + err.Error(),
			"link_title":  title,
			"url":         url,
			"description": description,
			"tags":        tagsStr,
		})
		return
	}

	// 处理标签
	if tagsStr != "" {
		// 分割标签字符串
		tagNames := strings.Split(tagsStr, ",")

		// 处理每个标签
		for _, name := range tagNames {
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
				c.HTML(http.StatusInternalServerError, "new_link", gin.H{
					"title":       "分享新链接 - LinkHUB",
					"error":       "处理标签失败: " + result.Error.Error(),
					"link_title":  title,
					"url":         url,
					"description": description,
					"tags":        tagsStr,
				})
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
				c.HTML(http.StatusInternalServerError, "new_link", gin.H{
					"title":       "分享新链接 - LinkHUB",
					"error":       "关联标签失败: " + err.Error(),
					"link_title":  title,
					"url":         url,
					"description": description,
					"tags":        tagsStr,
				})
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "new_link", gin.H{
			"title":       "分享新链接 - LinkHUB",
			"error":       "保存链接失败: " + err.Error(),
			"link_title":  title,
			"url":         url,
			"description": description,
			"tags":        tagsStr,
		})
		return
	}

	// 重定向到链接详情页
	c.Redirect(http.StatusFound, "/links/"+strconv.Itoa(int(link.ID)))
}

// ShowLink 显示链接详情页面
func ShowLink(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().Preload("User").Preload("Tags").First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"title": "链接不存在 - LinkHUB",
			"error": "链接不存在或已被删除",
		})
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
	var voted bool
	if userInterface, exists := c.Get("user"); exists {
		user := userInterface.(models.User)
		var count int64
		database.GetDB().Model(&models.Vote{}).Where("user_id = ? AND link_id = ?", user.ID, link.ID).Count(&count)
		voted = count > 0
	}

	// 渲染模板
	c.HTML(http.StatusOK, "link_detail", gin.H{
		"title":    link.Title + " - LinkHUB",
		"link":     link,
		"comments": comments,
		"voted":    voted,
		"userInfo": userinfo,
	})
}

// UpdateLink 更新链接处理函数
func UpdateLink(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	user := userInterface.(models.User)

	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", gin.H{
			"title": "链接不存在 - LinkHUB",
			"error": "链接不存在或已被删除",
		})
		return
	}

	// 验证用户权限
	if link.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"title": "没有权限 - LinkHUB",
			"error": "您没有权限编辑此链接",
		})
		return
	}

	// 获取表单数据
	title := c.PostForm("title")
	url := c.PostForm("url")
	description := c.PostForm("description")
	thumbnail := c.PostForm("thumbnail")
	tagsStr := c.PostForm("tags")

	// 验证表单数据
	if title == "" || url == "" {
		c.HTML(http.StatusBadRequest, "edit_link", gin.H{
			"title":       "编辑链接 - LinkHUB",
			"error":       "标题和URL是必填的",
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"thumbnail":   thumbnail,
			"tags":        tagsStr,
		})
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
		c.HTML(http.StatusInternalServerError, "edit_link", gin.H{
			"title":       "编辑链接 - LinkHUB",
			"error":       "更新链接失败: " + err.Error(),
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"thumbnail":   thumbnail,
			"tags":        tagsStr,
		})
		return
	}

	// 清除现有标签关联
	if err := tx.Model(&link).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "edit_link", gin.H{
			"title":       "编辑链接 - LinkHUB",
			"error":       "清除标签关联失败: " + err.Error(),
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"thumbnail":   thumbnail,
			"tags":        tagsStr,
		})
		return
	}

	// 处理标签
	if tagsStr != "" {
		// 分割标签字符串
		tagNames := strings.Split(tagsStr, ",")

		// 处理每个标签
		for _, name := range tagNames {
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
				c.HTML(http.StatusInternalServerError, "edit_link", gin.H{
					"title":       "编辑链接 - LinkHUB",
					"error":       "处理标签失败: " + result.Error.Error(),
					"link":        link,
					"link_title":  title,
					"url":         url,
					"description": description,
					"thumbnail":   thumbnail,
					"tags":        tagsStr,
				})
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
				c.HTML(http.StatusInternalServerError, "edit_link", gin.H{
					"title":       "编辑链接 - LinkHUB",
					"error":       "关联标签失败: " + err.Error(),
					"link":        link,
					"link_title":  title,
					"url":         url,
					"description": description,
					"thumbnail":   thumbnail,
					"tags":        tagsStr,
				})
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusInternalServerError, "edit_link", gin.H{
			"title":       "编辑链接 - LinkHUB",
			"error":       "保存链接失败: " + err.Error(),
			"link":        link,
			"link_title":  title,
			"url":         url,
			"description": description,
			"thumbnail":   thumbnail,
			"tags":        tagsStr,
		})
		return
	}

	// 重定向到链接详情页
	c.Redirect(http.StatusFound, "/links/"+strconv.Itoa(int(link.ID)))
}

// DeleteLink 删除链接处理函数
func DeleteLink(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	user := userInterface.(models.User)

	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "链接不存在或已被删除"})
		return
	}

	// 验证用户权限
	if link.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "您没有权限删除此链接"})
		return
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 清除标签关联
	if err := tx.Model(&link).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除标签关联失败: " + err.Error()})
		return
	}

	// 删除链接
	if err := tx.Delete(&link).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除链接失败: " + err.Error()})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除链接失败: " + err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "投票成功"})
}

// VoteLink 投票链接处理函数
func VoteLink(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)

	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}

	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.HTML(http.StatusOK, "result", gin.H{
			"title":         "Error",
			"message":       "链接不存在或已被删除",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 检查用户是否已投票
	var count int64
	database.GetDB().Model(&models.Vote{}).Where("user_id = ? AND link_id = ?", user.ID, link.ID).Count(&count)
	if count > 0 {
		c.HTML(http.StatusOK, "result", gin.H{
			"title":         "Warning",
			"message":       "您已经投过票了",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 创建投票记录
	vote := models.Vote{
		UserID: user.ID,
		LinkID: link.ID,
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 保存投票
	if err := tx.Create(&vote).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusOK, "result", gin.H{
			"title":         "Error",
			"message":       "投票失败: " + err.Error(),
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 更新链接投票计数
	if err := tx.Model(&link).Update("vote_count", link.VoteCount+1).Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusOK, "result", gin.H{
			"title":         "Error",
			"message":       "更新投票计数失败: " + err.Error(),
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusOK, "result", gin.H{
			"title":         "Error",
			"message":       "投票失败: " + err.Error(),
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}
	c.Redirect(302, refer)
}

// UnvoteLink 取消投票处理函数
func UnvoteLink(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	user := userInterface.(models.User)

	// 获取链接ID
	id := c.Param("id")

	// 查询链接
	var link models.Link
	result := database.GetDB().First(&link, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "链接不存在或已被删除"})
		return
	}

	// 检查用户是否已投票
	var vote models.Vote
	result = database.GetDB().Where("user_id = ? AND link_id = ?", user.ID, link.ID).First(&vote)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您还没有投过票"})
		return
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 删除投票记录
	if err := tx.Delete(&vote).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消投票失败: " + err.Error()})
		return
	}

	// 更新链接投票计数
	if err := tx.Model(&link).Update("vote_count", link.VoteCount-1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新投票计数失败: " + err.Error()})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消投票失败: " + err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "已取消投票"})
}
