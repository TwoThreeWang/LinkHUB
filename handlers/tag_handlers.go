package handlers

import (
	"github.com/gin-gonic/gin"
	"LinkHUB/database"
	"LinkHUB/models"
	"net/http"
)

// ListTags 获取标签列表
func ListTags(c *gin.Context) {
	// 获取所有标签
	var tags []models.Tag
	database.GetDB().Order("count DESC").Find(&tags)

	c.HTML(http.StatusOK, "tags", gin.H{
		"title": "标签列表 - LinkHUB",
		"tags":  tags,
	})
}

// ShowTag 显示标签详情
func ShowTag(c *gin.Context) {
	// 获取标签slug
	slug := c.Param("slug")

	// 查询标签
	var tag models.Tag
	result := database.GetDB().Where("slug = ?", slug).First(&tag)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"title": "标签不存在 - LinkHUB",
			"error": "标签不存在或已被删除",
		})
		return
	}

	// 获取标签下的链接
	var links []models.Link
	database.GetDB().Model(&tag).Association("Links").Find(&links)

	c.HTML(http.StatusOK, "tag_detail", gin.H{
		"title": tag.Name + " - LinkHUB",
		"tag":   tag,
		"links": links,
	})
}

// SuggestTags 标签建议API
func SuggestTags(c *gin.Context) {
	// 获取查询参数
	query := c.Query("q")

	// 搜索标签
	var tags []models.Tag
	database.GetDB().Where("name ILIKE ?", "%"+query+"%").Limit(5).Find(&tags)

	// 构建建议列表
	suggestions := make([]string, len(tags))
	for i, tag := range tags {
		suggestions[i] = tag.Name
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

// SearchLinks 搜索链接API
func SearchLinks(c *gin.Context) {
	// 获取查询参数
	query := c.Query("q")

	// 搜索链接
	var links []models.Link
	database.GetDB().Where(
		"title ILIKE ? OR description ILIKE ?",
		"%"+query+"%",
		"%"+query+"%",
	).Preload("User").Preload("Tags").Limit(10).Find(&links)

	c.JSON(http.StatusOK, gin.H{"links": links})
}
