package auth

import (
	"bookbox-backend/internal/config"
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/logger"
	"bookbox-backend/pkg/redis"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type SignedTokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type TokenPairCache struct {
	AccessTokenUUID  string    `json:"accessTokenUUID"`
	RefreshTokenUUID string    `json:"refreshTokenUUID"`
	EntryTime        time.Time `json:"entryTime"`
}

const (
	AccessTokenHeader  = "Auth-Access-Token"
	RefreshTokenHeader = "Auth-Refresh-Token"
)

func DestroyToken(ctx *gin.Context, userID string) error {
	err := redis.Client.HDel(userID, ctx.Request.Header.Get(RefreshTokenHeader)).Err()
	if err != nil {
		return err
	}

	return nil
}

func RefreshToken(ctx *gin.Context) (issuer *model.User, err error) {
	logger.Log.Info("validating refresh token")

	tokenToValidate := ctx.Request.Header.Get(RefreshTokenHeader)
	if err != nil {
		return nil, err
	}

	issuer, err = ValidateTokenClaims(ctx, tokenToValidate, true)
	if err != nil {
		return
	}

	err = NewAccessToken(ctx, issuer.ID)
	if err != nil {
		return
	}

	return
}

func NewTokenPair(ctx *gin.Context, userID string) error {
	refreshTokenUUID := uuid.New().String()
	accessTokenUUID := uuid.New().String()

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		Issuer:    userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JWT.RefreshExpiry)),
		ID:        refreshTokenUUID,
	}).
		SignedString(config.JWT.RefreshPrivateKey)
	if err != nil {
		return err
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		Issuer:    userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JWT.AccessExpiry)),
		ID:        accessTokenUUID,
	}).
		SignedString(config.JWT.AccessPrivateKey)
	if err != nil {
		return err
	}

	cacheJSON, err := json.Marshal(TokenPairCache{
		AccessTokenUUID:  accessTokenUUID,
		RefreshTokenUUID: refreshTokenUUID,
		EntryTime:        time.Now(),
	})
	if err != nil {
		return err
	}

	err = redis.Client.HSet(userID, refreshToken, string(cacheJSON)).Err()
	if err != nil {
		return err
	}

	ctx.Header(AccessTokenHeader, accessToken)
	ctx.Header(RefreshTokenHeader, refreshToken)
	return nil
}

// Used to refresh tokens.
func NewAccessToken(ctx *gin.Context, userID string) error {
	accessTokenUUID := uuid.New().String()

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		Issuer:    userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JWT.AccessExpiry)),
		ID:        accessTokenUUID,
	}).
		SignedString(config.JWT.AccessPrivateKey)
	if err != nil {
		return err
	}

	cacheJSON, err := redis.Client.HGet(userID, ctx.Request.Header.Get(RefreshTokenHeader)).
		Result()
	if err != nil {
		return err
	}

	cachedTokenPair := TokenPairCache{}
	err = json.Unmarshal([]byte(cacheJSON), &cachedTokenPair)
	if err != nil {
		return err
	}

	cachedTokenPair.AccessTokenUUID = accessTokenUUID
	if err != nil {
		return err
	}

	raw, err := json.Marshal(cachedTokenPair)
	if err != nil {
		return err
	}

	err = redis.Client.HSet(
		userID,
		ctx.Request.Header.Get(RefreshTokenHeader),
		string(raw),
	).Err()

	if err != nil {
		return err
	}

	ctx.Header(AccessTokenHeader, accessToken)
	ctx.Header(RefreshTokenHeader, ctx.Request.Header.Get(RefreshTokenHeader))

	return nil
}
