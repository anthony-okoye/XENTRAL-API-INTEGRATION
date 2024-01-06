package middlewares

import (
	"github.com/gin-gonic/gin"
)

var (
	Origin = "*"
)

// CORS will inject HTTP header Access-Control-Allow-Origin
func CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer middlewareRecovery()

		ctx.Writer.Header().Set("Access-Control-Allow-Origin", Origin)
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Auth-Access-Token, Auth-Refresh-Token")
		ctx.Writer.Header().Set("Access-Control-Expose-Headers", " Auth-Access-Token, Auth-Refresh-Token")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		ctx.Next()
	}
}
