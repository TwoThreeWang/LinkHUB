package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateCategory 新增分类
func CreateCategory(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.JSON(http.StatusOK, OutputApi(403, "请先登录"))
		return
	}
	if userInfo.Role != "admin" {
		c.JSON(http.StatusOK, OutputApi(403, "权限错误，非管理员无法创建分类！"))
		return
	}
	// 获取查询参数
	categoryName := c.Query("category")
	if categoryName == "" {
		c.JSON(http.StatusOK, OutputApi(400, "分类名称不能为空"))
		return
	}
	var category models.Category
	result := database.GetDB().Where("name = ?", categoryName).FirstOrCreate(&category, models.Category{
		Name: categoryName,
	}).Error

	if result != nil {
		c.JSON(http.StatusOK, OutputApi(400, "分类创建失败："+result.Error()))
		return
	}
	c.JSON(http.StatusOK, OutputApi(200, "分类创建成功"))
}

// ShowCategory 显示分类详情
func ShowCategory(c *gin.Context) {
	// 获取分类id
	id := c.Param("id")
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// 获取排序参数
	sort := c.DefaultQuery("sort", "new")

	// 查询分类
	var category models.Category
	result := database.GetDB().Where("id = ?", id).First(&category)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "分类不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取分类下的文章
	var articles []models.Article
	var total int64

	// 获取分页数据
	pageSize := 10
	offset := (page - 1) * pageSize

	// 构建查询
	query := database.GetDB().Model(&models.Article{})

	// 添加分类过滤
	query = query.Where("category_id = ?", id)

	// 添加排序
	switch sort {
	case "hot":
		query = query.Order("view_count DESC")
	default: // new
		query = query.Order("created_at DESC")
	}

	// 执行查询
	query.Count(&total)
	query.Offset(offset).Preload("User").Limit(pageSize).Find(&articles)

	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize

	c.HTML(http.StatusOK, "category_detail", OutputCommonSession(c, gin.H{
		"title":      category.Name,
		"category":   category,
		"articles":   articles,
		"page":       page,
		"totalPages": totalPages,
		"sort":       sort,
	}))
}

// UpdateCategory 更新分类
func UpdateCategory(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.JSON(http.StatusOK, OutputApi(403, "用户未登录"))
		return
	}
	if userInfo.Role != "admin" {
		c.JSON(http.StatusOK, OutputApi(403, "权限错误，非管理员无法更新分类"))
		return
	}

	// 获取分类ID和新名称
	id := c.Param("id")
	newName := c.Query("name")
	if newName == "" {
		c.JSON(http.StatusOK, OutputApi(400, "分类名称不能为空"))
		return
	}

	// 更新分类
	result := database.GetDB().Model(&models.Category{}).Where("id = ?", id).Update("name", newName)
	if result.Error != nil {
		c.JSON(http.StatusOK, OutputApi(400, "分类更新失败："+result.Error.Error()))
		return
	}

	c.JSON(http.StatusOK, OutputApi(200, "分类更新成功"))
}

// DeleteCategory 删除分类
func DeleteCategory(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.JSON(http.StatusOK, OutputApi(403, "用户未登录"))
		return
	}
	if userInfo.Role != "admin" {
		c.JSON(http.StatusOK, OutputApi(403, "权限错误，非管理员无法删除分类"))
		return
	}

	// 获取分类ID
	id := c.Param("id")

	// 检查分类是否存在
	var category models.Category
	if err := database.GetDB().First(&category, id).Error; err != nil {
		c.JSON(http.StatusOK, OutputApi(400, "分类不存在"))
		return
	}

	// 检查分类下是否有文章
	var count int64
	database.GetDB().Model(&models.Article{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, OutputApi(400, "该分类下还有文章，无法删除"))
		return
	}

	// 删除分类
	if err := database.GetDB().Delete(&category).Error; err != nil {
		c.JSON(http.StatusOK, OutputApi(400, "分类删除失败："+err.Error()))
		return
	}

	c.JSON(http.StatusOK, OutputApi(200, "分类删除成功"))
}
