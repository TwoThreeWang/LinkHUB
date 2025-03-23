package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListTags 获取标签列表
func ListTags(c *gin.Context) {
	// 获取所有标签
	var tags []models.Tag
	database.GetDB().Order("count DESC").Find(&tags)

	c.HTML(http.StatusOK, "tags", OutputCommonSession(c, gin.H{
		"title": "标签列表",
		"tags":  tags,
	}))
}

// CreateTag 新增标签
func CreateTag(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.JSON(http.StatusOK, OutputApi(403, "请先登录"))
		return
	}
	if userInfo.Role != "admin" {
		c.JSON(http.StatusOK, OutputApi(403, "权限错误，非管理员无法创建标签！"))
		return
	}
	// 获取查询参数
	tagName := c.Query("tag")
	if tagName == "" {
		c.JSON(http.StatusOK, OutputApi(403, "标签名称不能为空"))
		return
	}
	var tag models.Tag
	result := database.GetDB().Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{
		Name: tagName,
	}).Error

	if result != nil {
		c.JSON(http.StatusOK, OutputApi(400, "标签创建失败："+result.Error()))
		return
	}
	c.JSON(http.StatusOK, OutputApi(200, "标签创建成功"))
}

// ShowTag 显示标签详情
func ShowTag(c *gin.Context) {
	// 获取标签id
	id := c.Param("id")
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// 获取排序参数
	sort := c.DefaultQuery("sort", "new")

	// 查询标签
	var tag models.Tag
	result := database.GetDB().Where("id = ?", id).First(&tag)
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "标签不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}

	// 获取标签下的链接
	var links []models.Link
	var total int64

	// 获取分页数据
	pageSize := 10
	offset := (page - 1) * pageSize

	// 使用EXISTS子查询获取标签关联的链接
	// 构建查询
	query := database.GetDB().Model(&models.Link{}).
		Where("EXISTS (SELECT 1 FROM link_tags WHERE link_tags.link_id = links.id AND link_tags.tag_id = ?)", tag.ID)

	// 计算总数
	query.Count(&total)
	tag.Count = int(total)

	// 更新标签的count字段，使其与实际关联的链接数量一致
	if err := database.GetDB().Model(&models.Tag{}).Where("id = ?", tag.ID).Update("count", total).Error; err != nil {
		fmt.Println("更新标签的count字段失败：", err)
	}

	// 根据排序参数设置排序方式
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

	c.HTML(http.StatusOK, "tag_detail", OutputCommonSession(c, gin.H{
		"title":      tag.Name,
		"tag":        tag,
		"links":      links,
		"page":       page,
		"totalPages": totalPages,
		"sort":       sort,
	}))
}

// UpdateTag 更新标签
func UpdateTag(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.JSON(http.StatusOK, OutputApi(403, "请先登录"))
		return
	}
	if userInfo.Role != "admin" {
		c.JSON(http.StatusOK, OutputApi(403, "权限错误，非管理员无法创建标签！"))
		return
	}
	// 获取链接ID
	id := c.Param("id")
	// 获取查询参数
	tagName := c.Query("tag")
	if tagName == "" || id == "" || id == "undefined" {
		c.JSON(http.StatusOK, OutputApi(400, "参数错误"))
		return
	}
	var tag models.Tag
	result := database.GetDB().First(&tag, id)
	if result.Error != nil {
		c.JSON(http.StatusOK, OutputApi(400, "标签信息获取失败："+result.Error.Error()))
		return
	}
	tag.Name = tagName
	if err := database.GetDB().Save(&tag).Error; err != nil {
		c.JSON(http.StatusOK, OutputApi(400, "标签修改失败："+err.Error()))
		return
	}
	c.JSON(http.StatusOK, OutputApi(200, "标签修改成功"))
}

// DeleteTag 删除标签
func DeleteTag(c *gin.Context) {
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.JSON(http.StatusOK, OutputApi(403, "请先登录"))
		return
	}

	// 验证用户权限
	if userInfo.Role != "admin" {
		c.JSON(http.StatusOK, OutputApi(403, "权限错误，非管理员无法删除标签！"))
		return
	}

	// 获取链接ID
	id := c.Param("id")
	if id == "" || id == "undefined" {
		c.JSON(http.StatusOK, OutputApi(400, "参数错误"))
		return
	}
	// 查询标签
	var tag models.Tag
	result := database.GetDB().First(&tag, id)
	if result.Error != nil {
		c.JSON(http.StatusOK, OutputApi(400, "链接不存在或已被删除"))
		return
	}
	// 开始事务
	tx := database.GetDB().Begin()

	// 删除link_tags表中的关联关系
	if err := tx.Table("link_tags").Where("tag_id = ?", id).Delete(nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, OutputApi(400, "删除标签关系失败："+err.Error()))
		return
	}

	// 删除标签
	if err := tx.Delete(&tag).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, OutputApi(400, "删除标签失败："+err.Error()))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, OutputApi(400, "删除标签失败："+err.Error()))
		return
	}

	c.JSON(http.StatusOK, OutputApi(200, "标签删除成功"))
}
