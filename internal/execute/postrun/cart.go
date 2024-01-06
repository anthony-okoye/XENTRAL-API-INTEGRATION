package postrun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
)

func CartPostrunRead(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	cart := queried.(*model.Cart)
	defer func() {
		response = cart
	}()

	if request.Data.SalesChannelID == "" {
		return
	}

	salesChannel := model.SalesChannel{}
	err = database.DB.Find(&salesChannel, "id = ? OR domain = ?", request.Data.SalesChannelID, request.Data.SalesChannelID).Error
	if err != nil {
		return nil, err
	}

	for i, val := range cart.CartItems {
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
			cart.CartItems[i].Product.SellingPrice = scProducts.ChangedPrice
		}

		if scProducts.ChangedTitle != "" {
			cart.CartItems[i].Product.Title = scProducts.ChangedTitle
		}
	}

	return
}
