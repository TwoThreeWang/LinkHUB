package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

	// 查询文章
	var article models.Article
	result := database.GetDB().Preload("User").Preload("Category").First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "文章不存在",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取评论
	var comments []models.ArticleComment
	database.GetDB().Where("article_id = ? AND parent_id IS NULL", article.ID).
		Preload("User").
		Preload("Replies").
		Preload("Replies.User").
		Order("created_at ASC").
		Find(&comments)

	// 增加浏览量
	article.IncreaseViewCount()
	database.GetDB().Save(&article)
	// 推荐文章
	var relatedArticles []models.Article
	query := database.GetDB().Model(&models.Article{}).Where("category_id = ?", article.CategoryID).Where("id != ?", article.ID)

	// 按相关度（浏览量）和时间排序，限制5篇
	query.Order("view_count DESC, created_at DESC").Limit(5).Find(&relatedArticles)

	// 内容区广告
	contentAds := GetAdsByType(c, "content")
	sidebarAds := GetAdsByType(c, "sidebar")

	// 渲染模板
	c.HTML(http.StatusOK, "article_detail", OutputCommonSession(c, gin.H{
		"title":           article.Title,
		"article":         article,
		"comments":        comments,
		"relatedArticles": relatedArticles,
		"contentAds":      contentAds,
		"sidebarAds":      sidebarAds,
	}))
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
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
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
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "用户未登录",
			"redirect_text": "返回",
		}))
		return
	}

	// 查询文章
	var article models.Article
	result := database.GetDB().Preload("Category").First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "文章不存在",
			"redirect_text": "返回",
		}))
		return
	}

	// 检查权限
	if article.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "没有权限编辑此文章",
			"redirect_text": "返回",
		}))
		return
	}

	// 查询所有分类
	var categories []models.Category
	database.GetDB().Find(&categories)

	// 渲染模板
	c.HTML(http.StatusOK, "new_article", OutputCommonSession(c, gin.H{
		"title":      "编辑文章",
		"article":    article,
		"categories": categories,
	}))
}

// UpdateArticle 更新文章
func UpdateArticle(c *gin.Context) {
	// 获取文章ID
	id := c.Param("id")

	// 获取当前用户
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "用户未登录",
			"redirect_text": "返回",
		}))
		return
	}

	// 查询文章
	var article models.Article
	result := database.GetDB().First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "文章不存在",
			"redirect_text": "返回",
		}))
		return
	}

	// 检查权限
	if article.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "您没有权限编辑此文章",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取表单数据
	title := c.PostForm("title")
	content := c.PostForm("content")
	categoryID := c.PostForm("category")

	// 验证数据
	if title == "" || content == "" {
		// 查询所有分类
		var categories []models.Category
		database.GetDB().Find(&categories)
		c.HTML(http.StatusOK, "new_article", OutputCommonSession(c, gin.H{
			"title":      "编辑文章",
			"article":    article,
			"categories": categories,
			"error":      "标题和内容不能为空",
		}))
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
		c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "更新文章失败: " + result.Error.Error(),
			"redirect_text": "返回",
		}))
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
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "您没有权限删除此文章",
			"redirect_text": "返回",
		}))
		return
	}

	// 查询文章
	var article models.Article
	result := database.GetDB().First(&article, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "文章不存在",
			"redirect_text": "返回",
		}))
		return
	}

	// 检查权限
	if article.UserID != userInfo.ID && userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "您没有权限删除此文章",
			"redirect_text": "返回",
		}))
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
		c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "删除文章失败: " + result.Error.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 重定向到文章列表页
	c.Redirect(http.StatusFound, "/articles")
}

// SearchArticles 搜索文章
func SearchArticles(c *gin.Context) {
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
	// 获取文章列表
	var articles []models.Article
	var total int64

	// 构建查询
	queryDB := database.GetDB().Where(
		"title ILIKE ? OR content ILIKE ?", "%"+query+"%", "%"+query+"%",
	)
	// 先计算总记录数
	queryDB.Model(&models.Article{}).Count(&total)
	// 执行分页查询
	queryDB.Preload("Category").Preload("User").Limit(pageSize).Offset(offset).Find(&articles)
	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize

	// 渲染模板
	c.HTML(http.StatusOK, "article_search", OutputCommonSession(c, gin.H{
		"title":          query,
		"articles":       articles,
		"page":           page,
		"totalPages":     totalPages,
		"query":          query,
		"total":          total,
		"search_article": "search_article",
	}))
}
