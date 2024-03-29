package router

import (
	"bookbox-backend/internal/server/middlewares"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	Router = gin.New()
)

func init() {
	Router.Use(middlewares.NoCache())
	Router.Use(middlewares.Session())
	Router.Use(middlewares.CORS())
	Router.Use(middlewares.Security())

	middlewares.AuthKey = os.Getenv("AUTH_KEY")
	if middlewares.AuthKey != "" {
		Router.Use(middlewares.CheckAuthKey())
	}
}
