package routes

import (
	"LinkHUB/handlers"
	"LinkHUB/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine) {
	// 首页路由
	r.GET("/", handlers.Home)

	// 用户认证相关路由
	auth := r.Group("/auth")
	{
		auth.GET("/register", handlers.ShowRegister)
		auth.POST("/register", handlers.Register)
		auth.GET("/login", handlers.ShowLogin)
		auth.POST("/login", handlers.Login)
		auth.GET("/logout", middleware.AuthRequired(), handlers.Logout)
	}

	// 用户相关路由
	user := r.Group("/user")
	{
		user.GET("/profile", middleware.AuthRequired(), handlers.ShowProfile)
		user.GET("/profile/:id", handlers.ShowProfile)
		user.POST("/profile", middleware.AuthRequired(), handlers.UpdateProfile)
		user.GET("/links", middleware.AuthRequired(), handlers.UserLinks)
	}

	// 链接相关路由
	links := r.Group("/links")
	{
		links.GET("/", handlers.ListLinks)
		links.GET("/new", middleware.AuthRequired(), handlers.ShowNewLink)
		links.POST("/new", middleware.AuthRequired(), handlers.CreateLink)
		links.GET("/:id", handlers.ShowLink)
		links.PUT("/:id", middleware.AuthRequired(), handlers.UpdateLink)
		links.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteLink)
		links.GET("/:id/vote", middleware.AuthRequired(), handlers.VoteLink)
		links.GET("/:id/unvote", middleware.AuthRequired(), handlers.UnvoteLink)
	}

	// 评论相关路由
	comments := r.Group("/comments", middleware.AuthRequired())
	{
		comments.POST("/", handlers.CreateComment)
	}

	// 标签相关路由
	tags := r.Group("/tags")
	{
		tags.GET("/", handlers.ListTags)
		tags.GET("/:slug", handlers.ShowTag)
	}

	// API路由
	api := r.Group("/api")
	{
		api.GET("/tags/suggest", handlers.SuggestTags)
		api.GET("/links/search", handlers.SearchLinks)
	}
}
