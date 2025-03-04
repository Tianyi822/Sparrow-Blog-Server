package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"h2blog_server/pkg/config"
)

func ConfigServiceCors() gin.HandlerFunc {
	c := cors.Config{
		AllowAllOrigins: true, // 允许所有来源
		AllowMethods:    []string{"GET", "POST"},
		AllowHeaders:    []string{"Origin", "Content-Type"},
	}

	return cors.New(c)
}

func RunTimeCors() gin.HandlerFunc {
	c := cors.Config{
		AllowOrigins: config.Server.Cors.Origins,
		AllowMethods: config.Server.Cors.Methods,
		AllowHeaders: config.Server.Cors.Headers,
	}

	return cors.New(c)
}
