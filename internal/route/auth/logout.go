package auth

import (
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logout(ctx *gin.Context) {

	var (
		logoutResponse = request.Response{}
	)

	log := logger.Log.WithOptions(zap.Fields())

	log.Info("logout started")

	issuer, err := GetIssuer(ctx)
	if err != nil {
		log.Error("No issuer found",
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, logoutResponse, []string{err.Error()}, 400, log)
		return
	}

	err = DestroyToken(ctx, issuer.ID)
	if err != nil {
		log.Error("Failed to logout user",
			zap.Error(err),
		)

		fail.ReturnError(ctx, logoutResponse, []string{err.Error()}, 400, log)
		return
	}

	log.Info("logout finished")

	logoutResponse.Status = true
	ctx.JSON(200, logoutResponse)
}

func init() {
	router.Router.Handle("POST", "auth/logout", Logout)
}
