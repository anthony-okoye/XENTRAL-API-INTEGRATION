package prerun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"encoding/json"
	"fmt"
)

// OrderPrerunRead prerun functions for user
func OrderPrerunRead(req *request.GetRequest, issuer *model.User) (err error) {
	return
}

// OrderPrerunList prerun functions for user
func OrderPrerunList(req *request.GetRequest, issuer *model.User) (err error) {
	if issuer.Role == model.UserCustomerRole {
		req.Metadata.Filter.Must = append(req.Metadata.Filter.Must, request.FilterParam{
			Key:   "user_id",
			Value: issuer.ID,
			Type:  "eq",
		})
	}

	return
}

// OrderPrerunUpdate prerun functions for user
func OrderPrerunUpdate(request *request.Request, issuer *model.User) (err error) {
	request.Metadata.UpdateFields = []string{
		"order_status",
		"payment_status",
		"delivery_status",
		"payment_method",
		"active",
	}

	return
}

const (
	defaultOrderStatus    = "in_progress"
	defaultPaymentStatus  = "pending"
	defaultDeliveryStatus = "open"
)

// OrderPrerunCreate prerun functions for user
func OrderPrerunCreate(req *request.Request, issuer *model.User) (err error) {
	if issuer.Role != "admin" {
		req.Data["order_status"] = defaultOrderStatus
		req.Data["payment_status"] = defaultPaymentStatus
		req.Data["delivery_status"] = defaultDeliveryStatus
	}

	orderItems := []model.OrderItem{}
	raw, err := json.Marshal(req.Data["products"])
	if err != nil {
		return
	}

	salesChannelID, _ := req.Data["sales_channel_id"].(string)
	if salesChannelID == "" {
		err = fmt.Errorf("sales channel cannot be unspecified")
		return
	}

	json.Unmarshal(raw, &orderItems)
	if len(orderItems) == 0 {
		err = fmt.Errorf("no products present in order")
		return
	}

	totalPrice := float64(0)
	isAllDownloadTitle := true
	for i, orderItem := range orderItems {
		//check if product is valid and in stock
		if orderItem.ProductID == "" {
			err = fmt.Errorf("no product_id present in orderItem")
			return
		}

		row := model.Product{}
		res := database.DB.Find(&row, "id = ?", orderItem.ProductID)
		if res.Error != nil {
			return err
		}

		if res.RowsAffected == 0 {
			err = fmt.Errorf("that product doesn't exist")
			return
		}

		if !(*row.Active) {
			err = fmt.Errorf("invalid product found in order")
			return
		}

		if row.Stock == 0 {
			err = fmt.Errorf("product: %s is out of stock", row.Title)
			return
		}

		if row.Stock-orderItem.Quantity < 0 {
			err = fmt.Errorf("not enough product is available in stock")
			return
		}

		scProducts := &model.SalesChannelProduct{}
		res = database.DB.
			Find(scProducts, "product_id = ? and sales_channel_id = ?", row.ID, salesChannelID)
		if res.Error != nil {
			err = fmt.Errorf("specified sales channel doesn't exist")
			return
		}

		sellingPrice := row.SellingPrice
		if scProducts.ChangedPrice != 0 {
			sellingPrice = scProducts.ChangedPrice
		}

		if !row.IsDownloadTitle {
			isAllDownloadTitle = false
		}

		orderItems[i].CurrentPrice = sellingPrice
		totalPrice += (sellingPrice * float64(orderItem.Quantity))
	}

	if !isAllDownloadTitle && totalPrice < 40 {
		totalPrice += 3.99
	}

	req.Data["total_price"] = totalPrice
	req.Data["products"] = orderItems
	return
}

// OrderPrerunDelete prerun functions for user
func OrderPrerunDelete(req *request.Request, issuer *model.User) (err error) {
	return
}
