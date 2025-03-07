package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"LinkHUB/config"
	"LinkHUB/database"
	"LinkHUB/render"
	"LinkHUB/routes"
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

	// 使用多模板渲染器
	r.HTMLRender = render.CreateRenderer()

	// 设置静态文件目录
	r.Static("/static", "./static")

	// 注册路由
	routes.SetupRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.GetConfig().Server.Port)
	r.Run(addr)
}