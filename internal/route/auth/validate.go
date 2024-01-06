package auth

import (
	"bookbox-backend/internal/config"
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/logger"
	"bookbox-backend/pkg/redis"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

func GetIssuer(ctx *gin.Context) (issuer *model.User, err error) {
	logger.Log.Info("validating token")
	if ctx.Request.Header.Get(AccessTokenHeader) == "" && ctx.Request.Header.Get(RefreshTokenHeader) == "" {
		issuer = &model.User{
			Role: "guest",
		}
		return
	}

	issuer, err = ValidateTokenClaims(ctx, ctx.Request.Header.Get(AccessTokenHeader), false)
	if err != nil {
		return RefreshToken(ctx)
	}
	return
}

func ValidateTokenClaims(
	ctx *gin.Context,
	tokenToValidate string,
	isRefresh bool,
) (*model.User, error) {
	parsedToken, err := jwt.ParseWithClaims(
		tokenToValidate,
		&jwt.RegisteredClaims{},
		func(jwt *jwt.Token) (interface{}, error) {
			if isRefresh {
				return config.JWT.RefreshPrivateKey.Public(), nil
			} else {
				return config.JWT.AccessPrivateKey.Public(), nil
			}
		},
	)
	if err != nil {
		logger.Log.Warn("Failed to parse token data",
			zap.Error(err),
		)
		return nil, err
	}

	claims := parsedToken.Claims.(*jwt.RegisteredClaims)
	cacheJSON, err := redis.Client.HGet(claims.Issuer, ctx.Request.Header.Get(RefreshTokenHeader)).Result()
	if err != nil {
		logger.Log.Error("Failed to get data from Redis",
			zap.Error(err),
		)
		return nil, err
	}

	cachedTokenPair := TokenPairCache{}

	err = json.Unmarshal([]byte(cacheJSON), &cachedTokenPair)
	if err != nil {
		return nil, err
	}

	var tokenUID string
	if isRefresh {
		tokenUID = cachedTokenPair.RefreshTokenUUID
	} else {
		tokenUID = cachedTokenPair.AccessTokenUUID
	}

	if tokenUID != claims.ID {
		err = fmt.Errorf("token is invalid")
		return nil, err
	}

	var issuer model.User
	resp := database.DB.Where("id = ?", claims.Issuer).
		First(&issuer)

	if resp.RowsAffected < 1 || resp.Error != nil {
		err = fmt.Errorf("user doesn't exist in db")
		if resp.Error != nil {
			err = resp.Error
		}
		logger.Log.Error("Issuer not found",
			zap.Error(err),
		)

		return nil, err
	}

	return &issuer, nil
}
