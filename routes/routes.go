package routes

import (
	"LinkHUB/handlers"
	"LinkHUB/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine) {
	// 首页路由
	r.GET("/", handlers.Home) // 网站首页

	// 用户认证相关路由
	auth := r.Group("/auth")
	{
		auth.GET("/register", handlers.ShowRegister)                    // 用户注册
		auth.POST("/register", handlers.Register)                       // 用户注册处理逻辑
		auth.GET("/login", handlers.ShowLogin)                          // 用户登录
		auth.POST("/login", handlers.Login)                             // 用户登录处理逻辑
		auth.GET("/logout", middleware.AuthRequired(), handlers.Logout) // 退出登录
	}

	// 用户相关路由
	user := r.Group("/user")
	{
		user.GET("/profile", middleware.AuthRequired(), handlers.ShowProfile)    // 用户主页
		user.GET("/profile/:id", handlers.ShowProfile)                           // 指定ID用户主页
		user.POST("/profile", middleware.AuthRequired(), handlers.UpdateProfile) // 用户资料更新
	}

	// 链接相关路由
	links := r.Group("/links")
	{
		links.GET("/:id", handlers.ShowLink)                                         // 链接详情
		links.GET("/new", middleware.AuthRequired(), handlers.ShowNewLink)           // 新增链接
		links.POST("/new", middleware.AuthRequired(), handlers.CreateLink)           // 新增链接处理逻辑
		links.GET("/:id/update", middleware.AuthRequired(), handlers.ShowUpdateLink) // 修改链接
		links.POST("/:id/update", middleware.AuthRequired(), handlers.UpdateLink)    // 修改链接处理逻辑
		links.GET("/:id/delete", middleware.AuthRequired(), handlers.DeleteLink)     // 删除链接
		links.GET("/:id/vote", middleware.AuthRequired(), handlers.VoteLink)         // 链接投票
		links.GET("/:id/unvote", middleware.AuthRequired(), handlers.UnVoteLink)     // 取消投票
		links.GET("/:id/click", handlers.ClickLink)                                  // 点击链接
		links.GET("/search", handlers.SearchLinks)                                   // 搜索
	}

	// 评论相关路由
	comments := r.Group("/comments", middleware.AuthRequired())
	{
		comments.POST("/", handlers.CreateComment) // 创建链接评论
	}

	// 文章评论相关路由
	articleComments := r.Group("/article-comments", middleware.AuthRequired())
	{
		articleComments.POST("/", handlers.CreateArticleComment) // 创建文章评论
	}

	// 标签相关路由
	tags := r.Group("/tags")
	{
		tags.GET("/", handlers.ListTags)            // 所有标签
		tags.GET("/:id", handlers.ShowTag)          // 标签下链接
		tags.GET("/add", handlers.CreateTag)        // 创建链接
		tags.GET("/:id/update", handlers.UpdateTag) // 修改链接
		tags.GET("/:id/delete", handlers.DeleteTag) // 删除链接
	}

	// 文章相关路由
	articles := r.Group("/articles")
	{
		articles.GET("/", handlers.ListArticles)                                           // 文章列表
		articles.GET("/:id", handlers.ShowArticle)                                         // 文章详情
		articles.GET("/new", middleware.AuthRequired(), handlers.ShowNewArticle)           // 新增文章
		articles.POST("/new", middleware.AuthRequired(), handlers.CreateArticle)           // 新增文章处理逻辑
		articles.GET("/:id/update", middleware.AuthRequired(), handlers.ShowUpdateArticle) // 修改文章
		articles.POST("/:id/update", middleware.AuthRequired(), handlers.UpdateArticle)    // 修改文章处理逻辑
		articles.GET("/:id/delete", middleware.AuthRequired(), handlers.DeleteArticle)     // 删除文章
		articles.GET("/search", handlers.SearchArticles)                                   // 搜索文章
	}
}
