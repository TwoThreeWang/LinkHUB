package middleware

import (
	"LinkHUB/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired 用户认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从session中获取用户ID
		userinfo := handlers.GetCurrentUser(c)
		if userinfo == nil {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", userinfo)
		c.Next()
	}
}
