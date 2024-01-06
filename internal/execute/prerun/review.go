package prerun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"fmt"
)

// ReviewPrerunCreate prerun functions for user
func ReviewPrerunCreate(req *request.Request, issuer *model.User) (err error) {
	productID, ok := req.Data["product_id"].(string)
	if !ok {
		err = fmt.Errorf("product not specified")
		return
	}

	row := model.Review{}
	res := database.DB.Find(&row, "user_id = ? && product_id = ?", issuer.ID, productID)
	if res.Error != nil {
		return err
	}

	if res.RowsAffected != 0 {
		err = fmt.Errorf("you are only allowed one review per product")
		return
	}

	return
}
