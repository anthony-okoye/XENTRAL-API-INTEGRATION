package prerun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"fmt"
)

// UserPrerunRead prerun functions for user
func SalesChPrerunRead(request *request.GetRequest, issuer *model.User) (err error) {
	return
}

// UserPrerunList prerun functions for user
func SalesChPrerunList(request *request.GetRequest, issuer *model.User) (err error) {
	return
}

// UserPrerunUpdate prerun functions for user
func SalesChPrerunUpdate(request *request.Request, issuer *model.User) (err error) {
	name, _ := request.Data["name"].(string)
	domain, _ := request.Data["domain"].(string)
	id, _ := request.Data["id"].(string)

	return CheckIfSalesChannelExistsUpdate(id, name, domain)
}

// UserPrerunCreate prerun functions for user
func SalesChPrerunCreate(request *request.Request, issuer *model.User) (err error) {
	name, _ := request.Data["name"].(string)
	domain, _ := request.Data["domain"].(string)

	return CheckIfSalesChannelExists(name, domain)
}

func CheckIfSalesChannelExists(name string, sub string) (err error) {
	salesChannel := model.SalesChannel{}

	if name == "" {
		err = fmt.Errorf("salechannel name cannot be empty")
		return
	}

	isFound := database.DB.Where("name = ?", name).First(&salesChannel).Error
	if isFound == nil {
		err = fmt.Errorf("salechannel name already exists")
		return
	}

	if sub == "" {
		err = fmt.Errorf("salechannel domain cannot be empty")
		return
	}

	isFound = database.DB.Where("domain = ?", sub).First(&salesChannel).Error
	if isFound == nil {
		err = fmt.Errorf("salechannel domain already exists")
		return
	}

	return
}

func CheckIfSalesChannelExistsUpdate(id, name, sub string) (err error) {
	salesChannel := model.SalesChannel{}

	if name == "" {
		err = fmt.Errorf("salechannel name cannot be empty")
		return
	}

	isFound := database.DB.Where("name = ? and id != ?", name, id).First(&salesChannel).Error
	if isFound == nil {
		err = fmt.Errorf("salechannel name already exists")
		return
	}

	if sub == "" {
		err = fmt.Errorf("salechannel domain cannot be empty")
		return
	}

	isFound = database.DB.Where("domain = ? and id != ?", sub, id).First(&salesChannel).Error
	if isFound == nil {
		err = fmt.Errorf("salechannel domain already exists")
		return
	}

	return
}

// UserPrerunDelete prerun functions for user
func SalesChPrerunDelete(request *request.Request, issuer *model.User) (err error) {
	if val, ok := request.Data["id"].(string); ok {
		if val == "1" {
			err = fmt.Errorf("unable to remove root sale channel")
			return
		}

	}

	return
}
