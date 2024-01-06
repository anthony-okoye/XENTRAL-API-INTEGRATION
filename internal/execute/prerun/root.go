package prerun

import (
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"

	"go.uber.org/zap"
)

var (
	ReadFunctions = map[string]func(*request.GetRequest, *model.User) error{
		"user":          UserPrerunRead,
		"order":         OrderPrerunRead,
		"product":       ProductPrerunRead,
		"sales_channel": SalesChPrerunRead,
	}

	ListFunctions = map[string]func(*request.GetRequest, *model.User) error{
		"user":          UserPrerunList,
		"order":         OrderPrerunList,
		"product":       ProductPrerunList,
		"sales_channel": SalesChPrerunList,
	}

	CreateFunctions = map[string]func(*request.Request, *model.User) error{
		"user":          UserPrerunCreate,
		"order":         OrderPrerunCreate,
		"product":       ProductPrerunCreate,
		"sales_channel": SalesChPrerunCreate,
		"review":        ReviewPrerunCreate,
	}

	DeleteFunctions = map[string]func(*request.Request, *model.User) error{
		"user":          UserPrerunDelete,
		"order":         OrderPrerunDelete,
		"product":       ProductPrerunDelete,
		"sales_channel": SalesChPrerunDelete,
	}

	UpdateFunctions = map[string]func(*request.Request, *model.User) error{
		"user":          UserPrerunUpdate,
		"order":         OrderPrerunUpdate,
		"product":       ProductPrerunUpdate,
		"sales_channel": SalesChPrerunUpdate,
	}

	CacheFunctions = map[string]func(request.GetRequest, *model.User, *zap.Logger) (request.Response, bool, error){
		"product":  ProductPrerunCache,
		"category": CategoryPrerunCache,
	}
)
