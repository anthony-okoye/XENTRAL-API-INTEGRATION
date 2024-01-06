package prerun

import (
	"bookbox-backend/internal/execute/validation"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
)

// UserPrerunRead prerun functions for user
func UserPrerunRead(request *request.GetRequest, issuer *model.User) (err error) {
	// if user role is customer or if nothing is specified load yourself
	if issuer.Role == model.UserCustomerRole || request.Data.ID == "" {
		request.Data.ID = issuer.ID
	}

	return
}

// UserPrerunList prerun functions for user
func UserPrerunList(request *request.GetRequest, issuer *model.User) (err error) {
	return
}

// UserPrerunUpdate prerun functions for user
func UserPrerunUpdate(request *request.Request, issuer *model.User) (err error) {
	_, ok := request.Data["id"].(string)
	if issuer.Role == model.UserCustomerRole || !ok {
		request.Data["role"] = model.UserCustomerRole
		request.Data["id"] = issuer.ID
		delete(request.Data, "orders")
		delete(request.Data, "reviews")
		delete(request.Data, "discounts")
	}

	return validation.UserValidateUpdate(request, issuer)
}

// UserPrerunCreate prerun functions for user
func UserPrerunCreate(request *request.Request, issuer *model.User) (err error) {
	// init user role
	if issuer.Role != model.UserAdminRole {
		if request.Data == nil {
			request.Data = make(map[string]any)
		}

		request.Data["role"] = model.UserCustomerRole
	}

	err = validation.UserValidateCreate(request, issuer)
	if err != nil {
		return
	}

	return
}

// UserPrerunDelete prerun functions for user
func UserPrerunDelete(request *request.Request, issuer *model.User) (err error) {
	return
}
