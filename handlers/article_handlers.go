package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListArticles 获取文章列表
func ListArticles(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// 获取排序参数
	sort := c.DefaultQuery("sort", "new")

	// 获取文章列表
	var articles []models.Article
	var total int64

	// 构建查询
	query := database.GetDB().Model(&models.Article{})

	// 计算总数
	query.Count(&total)

	// 获取分页数据
	pageSize := 10
	offset := (page - 1) * pageSize

	// 根据排序参数设置排序方式
	switch sort {
	case "top":
		// 使用like_count和view_count的和进行排序
		query = query.Order("view_count DESC")
	case "new":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// 执行查询
	query.Limit(pageSize).
		Offset(offset).
		Preload("User").
		Preload("Category").
		Find(&articles)

	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize

	// 获取所有分类
	var categories []models.Category
	database.GetDB().Order("count DESC").Find(&categories)

	// 渲染模板
	c.HTML(http.StatusOK, "articles", OutputCommonSession(c, gin.H{
		"title":      "文章列表",
		"articles":   articles,
		"page":       page,
		"totalPages": totalPages,
		"sort":       sort,
		"categories": categories,
	}))
}

// ShowArticle 显示文章详情
func ShowArticle(c *gin.Context) {
	// 获取文章ID
	id := c.Param("id")

	// 获取当前用户
	userInfo := GetCurrentUser(c)

	// 查询文章
	var article models.Article
	result := database.GetDB().Preload("User").Preload("Category").Preload("Comments").Preload("Comments.User").Preload("Comments.Replies").Preload("Comments.Replies.User").First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "文章不存在",
			"redirect_text": "返回首页",
			"redirect_url":  "/",
		})
		return
	}

	// 增加浏览量
	article.IncreaseViewCount()
	database.GetDB().Save(&article)

	// 渲染模板
	c.HTML(http.StatusOK, "article_detail", gin.H{
		"title":    article.Title,
		"article":  article,
		"userInfo": userInfo,
	})
}

// ShowNewArticle 显示创建文章页面
func ShowNewArticle(c *gin.Context) {
	// 获取当前用户
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "请先登录",
			"redirect_text": "去登陆",
			"refer":         "/auth/login",
		}))
		return
	}

	// 查询所有分类
	var categories []models.Category
	database.GetDB().Find(&categories)

	// 渲染模板
	c.HTML(http.StatusOK, "new_article", OutputCommonSession(c, gin.H{
		"title":      "创建文章",
		"categories": categories,
	}))
}

// CreateArticle 创建文章
func CreateArticle(c *gin.Context) {
	// 获取当前用户
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "请先登录",
			"redirect_text": "去登陆",
			"refer":         "/auth/login",
		}))
		return
	}

	// 获取表单数据
	title := c.PostForm("title")
	content := c.PostForm("content")
	categoryID := c.PostForm("category")

	// 查询所有分类
	var categories []models.Category
	database.GetDB().Find(&categories)

	// 验证数据
	if title == "" || content == "" {
		c.HTML(http.StatusOK, "new_article", OutputCommonSession(c, gin.H{
			"title":      "创建文章",
			"error":      "标题和内容不能为空",
			"categories": categories,
			"posttitle":  title,
			"content":    content,
			"categoryID": categoryID,
		}))
		return
	}

	// 创建文章
	article := models.Article{
		Title:   title,
		Content: content,
		UserID:  userInfo.ID,
	}

	// 开始数据库事务
	tx := database.GetDB().Begin()

	// 设置分类
	if categoryID != "" {
		catID, err := strconv.Atoi(categoryID)
		if err == nil {
			article.CategoryID = uint(catID)

			// 增加分类计数
			var category models.Category
			if tx.First(&category, catID).Error == nil {
				category.IncreaseCount()
				tx.Save(&category)
			}
		}
	}

	// 保存文章
	result := tx.Create(&article)
	if result.Error != nil {
		tx.Rollback()
		c.HTML(http.StatusOK, "new_article", OutputCommonSession(c, gin.H{
			"title":      "创建文章",
			"error":      "创建文章失败: " + result.Error.Error(),
			"categories": categories,
			"posttitle":  title,
			"content":    content,
			"categoryID": categoryID,
		}))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.HTML(http.StatusOK, "new_article", OutputCommonSession(c, gin.H{
			"title":      "创建文章",
			"error":      "保存文章失败: " + err.Error(),
			"categories": categories,
			"posttitle":  title,
			"content":    content,
			"categoryID": categoryID,
		}))
		return
	}

	// 重定向到文章详情页
	c.Redirect(http.StatusFound, "/articles/"+strconv.Itoa(int(article.ID)))
}

// ShowUpdateArticle 显示更新文章页面
func ShowUpdateArticle(c *gin.Context) {
	// 获取文章ID
	id := c.Param("id")

	// 获取当前用户
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	// 查询文章
	var article models.Article
	result := database.GetDB().Preload("Category").First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "文章不存在",
			"redirect_text": "返回首页",
			"redirect_url":  "/",
		})
		return
	}

	// 检查权限
	if article.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "您没有权限编辑此文章",
			"redirect_text": "返回首页",
			"redirect_url":  "/",
		})
		return
	}

	// 查询所有分类
	var categories []models.Category
	database.GetDB().Find(&categories)

	// 渲染模板
	c.HTML(http.StatusOK, "update_article", gin.H{
		"title":      "编辑文章",
		"article":    article,
		"categories": categories,
		"userInfo":   userInfo,
	})
}

// UpdateArticle 更新文章
func UpdateArticle(c *gin.Context) {
	// 获取文章ID
	id := c.Param("id")

	// 获取当前用户
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	// 查询文章
	var article models.Article
	result := database.GetDB().First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "文章不存在",
			"redirect_text": "返回首页",
			"redirect_url":  "/",
		})
		return
	}

	// 检查权限
	if article.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "您没有权限编辑此文章",
			"redirect_text": "返回首页",
			"redirect_url":  "/",
		})
		return
	}

	// 获取表单数据
	title := c.PostForm("title")
	content := c.PostForm("content")
	categoryID := c.PostForm("category")

	// 验证数据
	if title == "" || content == "" {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "标题和内容不能为空",
			"redirect_text": "返回",
			"redirect_url":  "/articles/" + id + "/update",
		})
		return
	}

	// 如果分类发生变化，更新分类计数
	if article.CategoryID > 0 && article.CategoryID != 0 {
		// 减少原分类计数
		var oldCategory models.Category
		if database.GetDB().First(&oldCategory, article.CategoryID).Error == nil {
			oldCategory.DecreaseCount()
			database.GetDB().Save(&oldCategory)
		}
	}

	// 更新文章
	article.Title = title
	article.Content = content

	// 设置新分类
	if categoryID != "" {
		catID, err := strconv.Atoi(categoryID)
		if err == nil {
			article.CategoryID = uint(catID)

			// 增加新分类计数
			var category models.Category
			if database.GetDB().First(&category, catID).Error == nil {
				category.IncreaseCount()
				database.GetDB().Save(&category)
			}
		}
	} else {
		article.CategoryID = 0
	}

	// 保存文章
	result = database.GetDB().Save(&article)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "更新文章失败: " + result.Error.Error(),
			"redirect_text": "返回",
			"redirect_url":  "/articles/" + id + "/update",
		})
		return
	}

	// 重定向到文章详情页
	c.Redirect(http.StatusFound, "/articles/"+id)
}

// DeleteArticle 删除文章
func DeleteArticle(c *gin.Context) {
	// 获取文章ID
	id := c.Param("id")

	// 获取当前用户
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	// 查询文章
	var article models.Article
	result := database.GetDB().First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "文章不存在",
			"redirect_text": "返回首页",
			"redirect_url":  "/",
		})
		return
	}

	// 检查权限
	if article.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "您没有权限删除此文章",
			"redirect_text": "返回首页",
			"redirect_url":  "/",
		})
		return
	}

	// 如果文章有分类，减少分类计数
	if article.CategoryID > 0 {
		var category models.Category
		if database.GetDB().First(&category, article.CategoryID).Error == nil {
			category.DecreaseCount()
			database.GetDB().Save(&category)
		}
	}

	// 删除文章
	result = database.GetDB().Delete(&article)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "result", gin.H{
			"userInfo":      userInfo,
			"title":         "错误",
			"message":       "删除文章失败: " + result.Error.Error(),
			"redirect_text": "返回",
			"redirect_url":  "/articles",
		})
		return
	}

	// 重定向到文章列表页
	c.Redirect(http.StatusFound, "/articles")
}

// SearchArticles 搜索文章
func SearchArticles(c *gin.Context) {
	userInfo := GetCurrentUser(c)
	// 获取搜索关键词
	keyword := c.Query("q")

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// 获取排序参数
	sort := c.DefaultQuery("sort", "new")

	// 获取文章列表
	var articles []models.Article
	var total int64

	// 构建查询
	query := database.GetDB().Model(&models.Article{})

	// 添加搜索条件
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	query.Count(&total)

	// 获取分页数据
	pageSize := 10
	offset := (page - 1) * pageSize

	// 根据排序参数设置排序方式
	switch sort {
	case "top":
		query = query.Order("view_count DESC")
	case "new":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// 执行查询
	query.Limit(pageSize).
		Offset(offset).
		Preload("User").
		Preload("Category").
		Find(&articles)

	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize

	// 渲染模板
	c.HTML(http.StatusOK, "search", gin.H{
		"title":      "搜索结果: " + keyword,
		"articles":   articles,
		"page":       page,
		"totalPages": totalPages,
		"sort":       sort,
		"keyword":    keyword,
		"total":      total,
		"userInfo":   userInfo,
	})
}
