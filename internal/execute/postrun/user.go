package postrun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
)

func UserPostrunRead(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	user := queried.(*model.User)
	defer func() {
		response = user
	}()

	for j, val := range user.Favorites {
		scProducts := &model.SalesChannelProduct{}
		res := database.DB.
			Find(scProducts, "product_id = ? and sales_channel_id = ?", val.ProductID, val.SalesChannelID)
		if res.Error != nil {
			return
		}

		if res.RowsAffected == 0 {
			return
		}

		if scProducts.ChangedPrice != 0 {
			user.Favorites[j].Product.SellingPrice = scProducts.ChangedPrice
		}

		if scProducts.ChangedTitle != "" {
			user.Favorites[j].Product.Title = scProducts.ChangedTitle
		}
	}

	return
}

func UserPostrunList(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	users := queried.(*[]model.User)
	defer func() {
		response = users
	}()

	for i := range *users {
		for j, val := range (*users)[i].Favorites {
			scProducts := &model.SalesChannelProduct{}
			res := database.DB.
				Find(scProducts, "product_id = ? and sales_channel_id = ?", val.ProductID, val.SalesChannelID)
			if res.Error != nil {
				return
			}

			if res.RowsAffected == 0 {
				return
			}

			if scProducts.ChangedPrice != 0 {
				(*users)[i].Favorites[j].Product.SellingPrice = scProducts.ChangedPrice
			}

			if scProducts.ChangedTitle != "" {
				(*users)[i].Favorites[j].Product.Title = scProducts.ChangedTitle
			}
		}
	}

	return
}
