package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 跨域中间件
func Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	// 重新设置默认设置中的跨域字段
	config.AllowMethods = []string{"GET", "POST"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Cookie"}
	config.AllowOrigins = []string{"*"}
	config.AllowCredentials = true
	return cors.New(config)
}
