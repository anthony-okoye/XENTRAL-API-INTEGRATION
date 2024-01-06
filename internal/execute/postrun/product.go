package postrun

import (
	"bookbox-backend/internal/cache"
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"encoding/json"

	"go.uber.org/zap"
)

func ProductPostrunRead(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	product := queried.(*model.Product)
	defer func() {
		response = product
	}()

	if request.Data.SalesChannelID == "" {
		return
	}

	salesChannel := model.SalesChannel{}
	err = database.DB.Find(&salesChannel, "id = ? OR domain = ?", request.Data.SalesChannelID, request.Data.SalesChannelID).Error
	if err != nil {
		return nil, err
	}

	scProducts := &model.SalesChannelProduct{}
	res := database.DB.
		Find(scProducts, "product_id = ? and sales_channel_id = ?", product.ID, salesChannel.ID)
	if res.Error != nil {
		return
	}

	if res.RowsAffected == 0 {
		return
	}

	if scProducts.ChangedPrice != 0 {
		product.SellingPrice = scProducts.ChangedPrice
	}

	if scProducts.ChangedTitle != "" {
		product.Title = scProducts.ChangedTitle
	}

	return
}

func ProductPostrunList(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	products := queried.(*[]model.Product)
	defer func() {
		response = products
	}()

	if request.Data.SalesChannelID == "" {
		return
	}

	salesChannel := model.SalesChannel{}
	err = database.DB.Find(&salesChannel, "id = ? OR domain = ?", request.Data.SalesChannelID, request.Data.SalesChannelID).Error
	if err != nil {
		return nil, err
	}

	for i, val := range *products {
		scProducts := &model.SalesChannelProduct{}
		res := database.DB.
			Find(scProducts, "product_id = ? and sales_channel_id = ?", val.ID, salesChannel.ID)
		if res.Error != nil {
			return
		}

		if res.RowsAffected == 0 {
			return
		}

		if scProducts.ChangedPrice != 0 {
			(*products)[i].SellingPrice = scProducts.ChangedPrice
		}

		if scProducts.ChangedTitle != "" {
			(*products)[i].Title = scProducts.ChangedTitle
		}
	}

	return
}

func ProductListCache(request request.GetRequest, response request.Response, issuer *model.User, log *zap.Logger) (err error) {
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

func ProductCreateCache(request request.Request, response request.Response, issuer *model.User, log *zap.Logger) (err error) {
	cache.DataCache.DeleteAll(request.Entity + ":")
	log.Info("deleted cached data")

	return
}

func ProductUpdateCache(request request.Request, response request.Response, issuer *model.User, log *zap.Logger) (err error) {
	cache.DataCache.DeleteAll(request.Entity + ":")
	log.Info("deleted cached data")

	return
}
