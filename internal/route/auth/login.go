package auth

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/crypto"
	"bookbox-backend/pkg/logger"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Login(ctx *gin.Context) {
	var (
		loginRequest    = request.Request{}
		loginResponse   = request.Response{}
		email, password string
		ok              bool
	)

	// Parse the posted JSON data.
	err := ctx.ShouldBind(&loginRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, loginResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.Any("data", loginRequest.Data),
	))

	log.Info("login started")

	if email, ok = loginRequest.Data["email"].(string); !ok {
		err := fmt.Errorf("email is missing")
		log.Error("auth/login failed")

		fail.ReturnError(ctx, loginResponse, []string{err.Error()}, 400, log)
		return
	}

	if password, ok = loginRequest.Data["password"].(string); !ok {
		err := fmt.Errorf("password is missing")
		log.Error("auth/login failed")

		fail.ReturnError(ctx, loginResponse, []string{err.Error()}, 400, log)
		return
	}

	user, err := CheckPassword(email, password)
	if err != nil {
		log.Error("auth/login failed",
			zap.Error(err),
		)

		fail.ReturnError(ctx, loginResponse, []string{err.Error()}, 400, log)
		return
	}

	// Generate JWT pair.
	err = NewTokenPair(ctx, user.ID)
	if err != nil {
		log.Error("Failed to generate token",
			zap.Error(err),
		)

		fail.ReturnError(ctx, loginResponse, []string{err.Error()}, 400, log)
		return
	}

	log.Info("login finished")

	loginResponse.Status = true
	ctx.JSON(200, loginResponse)
}

func CheckPassword(email string, password string) (user model.User, err error) {
	// Look for the provided user.
	err = database.DB.Select("id", "password").Where("email = ?", email).First(&user).Error
	// User not found.
	if err != nil {
		return
	}

	splits := strings.Split(user.Password, ".")
	salt := splits[0]
	password = fmt.Sprintf("%s.%s", salt, password)
	password = crypto.SHA256(password)
	if len(splits) < 2 || password != splits[1] {
		err = fmt.Errorf("username or password are incorrect")
		return
	}

	return
}

func init() {
	router.Router.Handle("POST", "auth/login", Login)
}
