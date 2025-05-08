package main

import (
	"LinkHUB/handlers"
	"LinkHUB/utils"
	"fmt"
	"github.com/gin-contrib/gzip"

	"LinkHUB/config"
	"LinkHUB/database"
	"LinkHUB/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化数据库连接
	if err := database.InitDB(); err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	// 设置gin模式
	gin.SetMode(config.GetConfig().Server.Mode)

	// 创建gin实例
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	// 使用多模板渲染器
	r.HTMLRender = handlers.LoadLocalTemplates("./templates")

	// 设置静态文件目录
	r.Static("/static", "./static")

	// 设置robots.txt直接访问
	r.StaticFile("/robots.txt", "./static/robots.txt")
	// 设置静态文件直接访问
	r.StaticFile("/hahahaha.ads.controller.js", "./static/js/hahahaha.ads.controller.js")
	// 将缓存实例绑定到 Gin 上下文
	r.Use(func(c *gin.Context) {
		c.Set("cache", utils.GlobalCache)
		c.Next()
	})
	// 注册路由
	routes.SetupRoutes(r)
	// 处理 404 错误
	r.NoRoute(func(c *gin.Context) {
		c.HTML(404, "result", handlers.OutputCommonSession(c, gin.H{
			"title":         "404",
			"message":       "页面没有找到或已被移除，请请返回首页继续浏览！",
			"refer":         "/",
			"redirect_text": "返回首页",
		}))
	})
	// 启动服务器
	addr := fmt.Sprintf(":%d", config.GetConfig().Server.Port)
	r.Run(addr)
}
