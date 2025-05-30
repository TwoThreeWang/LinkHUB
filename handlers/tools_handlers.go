package handlers

import (
	"LinkHUB/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ToolsHome 工具页面
func ToolsHome(c *gin.Context) {
	c.HTML(http.StatusOK, "tools", OutputCommonSession(c, gin.H{
		"title": "在线工具",
	}))
}

// ArticleInsightAiTools AI文章总结工具页面
func ArticleInsightAiTools(c *gin.Context) {
	c.HTML(http.StatusOK, "article_insight_ai", OutputCommonSession(c, gin.H{
		"title": "AI文章总结工具",
	}))
}

// HtmlRunTools Html在线运行
func HtmlRunTools(c *gin.Context) {
	c.HTML(http.StatusOK, "tool_html_run", OutputCommonSession(c, gin.H{
		"title": "Html在线运行测试工具",
	}))
}

// MdEditTools 在线MD编辑器
func MdEditTools(c *gin.Context) {
	c.HTML(http.StatusOK, "tool_md_edit", OutputCommonSession(c, gin.H{
		"title": "在线MarkDown编辑器",
	}))
}

// ClearCache 清除所有缓存
func ClearCache(c *gin.Context) {
	// 重新初始化全局缓存
	utils.GlobalCache = utils.NewCache()

	c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
		"title":         "缓存清理",
		"message":       "缓存清理成功",
		"redirect_text": "返回",
	}))
	return
}