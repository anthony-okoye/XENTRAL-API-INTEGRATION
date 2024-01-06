package postrun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
)

func SalesChannelPostrunRead(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	salesChannel := queried.(*model.SalesChannel)
	defer func() {
		response = salesChannel
	}()

	for i, val := range salesChannel.Products {
		scProducts := &model.SalesChannelProduct{}
		res := database.DB.
			Find(scProducts, "product_id = ? and sales_channel_id = ?", val.ProductID, salesChannel.ID)
		if res.Error != nil {
			return
		}

		if res.RowsAffected == 0 {
			return
		}

		if scProducts.ChangedPrice != 0 {
			salesChannel.Products[i].Product.SellingPrice = scProducts.ChangedPrice
		}

		if scProducts.ChangedTitle != "" {
			salesChannel.Products[i].Product.Title = scProducts.ChangedTitle
		}
	}

	return
}

func SalesChannelPostrunList(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	salesChannels := queried.(*[]model.SalesChannel)
	defer func() {
		response = salesChannels
	}()

	if request.Data.SalesChannelID == "" {
		return
	}

	salesChannel := model.SalesChannel{}
	err = database.DB.Find(&salesChannel, "id = ? OR domain = ?", request.Data.SalesChannelID, request.Data.SalesChannelID).Error
	if err != nil {
		return nil, err
	}

	for i := range *salesChannels {
		for j, val := range (*salesChannels)[i].Products {
			scProducts := &model.SalesChannelProduct{}
			res := database.DB.
				Find(scProducts, "product_id = ? and sales_channel_id = ?", val.ProductID, (*salesChannels)[i].ID)
			if res.Error != nil {
				return
			}

			if res.RowsAffected == 0 {
				return
			}

			if scProducts.ChangedPrice != 0 {
				(*salesChannels)[i].Products[j].Product.SellingPrice = scProducts.ChangedPrice
			}

			if scProducts.ChangedTitle != "" {
				(*salesChannels)[i].Products[j].Product.Title = scProducts.ChangedTitle
			}
		}
	}

	return
}
