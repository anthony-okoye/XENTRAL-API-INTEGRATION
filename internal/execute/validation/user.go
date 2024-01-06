package validation

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/auth"
	"bookbox-backend/pkg/crypto"
	"fmt"
	"net/mail"
)

func UserValidateUpdate(request *request.Request, issuer *model.User) (err error) {

	password, _ := request.Data["password"].(string)
	oldPassword, _ := request.Data["old_password"].(string)
	id, _ := request.Data["id"].(string)

	if password != "" || oldPassword != "" {
		if issuer.ID != id {
			err = fmt.Errorf("can only edit your own password")
			return
		}

		_, err = auth.CheckPassword(issuer.Email, oldPassword)
		if err != nil {
			err = fmt.Errorf("old_password is not correct")
			return
		}

		if password != "" {
			request.Data["password"] = crypto.HashPassword(password)
		}

	} else {
		//default to old password
		user := model.User{}
		err = database.DB.Where("id = ?", id).First(&user).Error
		if err != nil {
			err = fmt.Errorf("user with specified id does not exist")
			return
		}

		request.Data["password"] = user.Password
	}

	err = CheckIfExists(request)

	if err != nil {
		return
	}

	return nil
}

func UserValidateCreate(request *request.Request, issuer *model.User) (err error) {
	password, _ := request.Data["password"].(string)
	if len(password) == 0 {
		err = fmt.Errorf("password cannot be empty")
		return
	}

	err = CheckIfExists(request)
	if err != nil {
		return
	}

	if password != "" {
		request.Data["password"] = crypto.HashPassword(password)
	}

	return
}

// CheckIfExists checks if user exists
func CheckIfExists(createRequest *request.Request) (err error) {
	var (
		eMail string
		ok    bool
		user  model.User
	)
	if eMail, ok = createRequest.Data["email"].(string); !ok {
		return
	}

	_, err = mail.ParseAddress(eMail)
	if err != nil {
		err = fmt.Errorf("email is not in a valid format")
		return
	}

	// Look for the provided user.
	isFound := database.DB.Where("email = ?", eMail).First(&user).Error
	// User not found.

	if isFound == nil && user.ID != createRequest.Data["id"] {
		err = fmt.Errorf("user with received email already exists")
		return
	}

	return
}
