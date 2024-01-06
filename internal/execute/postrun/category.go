package postrun

import (
	"bookbox-backend/internal/cache"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"encoding/json"

	"go.uber.org/zap"
)

func CategoryListCache(request request.GetRequest, response request.Response, issuer *model.User, log *zap.Logger) (err error) {
	var (
		raw []byte
	)

	raw, err = json.Marshal(request)
	if err != nil {
		return
	}

	Cache(raw, request.Entity, response)
	log.Info("saved cache data for request",
		zap.Any("request", request),
	)

	return
}

func CategoryCreateCache(request request.Request, response request.Response, issuer *model.User, log *zap.Logger) (err error) {
	cache.DataCache.DeleteAll(request.Entity + ":")
	log.Info("deleted cached data")

	return
}

func CategoryUpdateCache(request request.Request, response request.Response, issuer *model.User, log *zap.Logger) (err error) {
	cache.DataCache.DeleteAll(request.Entity + ":")
	log.Info("deleted cached data")

	return
}
