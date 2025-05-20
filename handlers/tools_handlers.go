package handlers

import (
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