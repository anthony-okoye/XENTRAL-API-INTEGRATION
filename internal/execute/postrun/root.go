package postrun

import (
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"

	"go.uber.org/zap"
)

var (
	CreateFunctions = map[string]func(any, *model.User) (any, error){
		"order": OrderPostrunCreate,
	}

	ReadFunctions = map[string]func(request.GetRequest, any, *model.User) (any, error){
		"user":          UserPostrunRead,
		"product":       ProductPostrunRead,
		"order":         OrderPostrunRead,
		"cart":          CartPostrunRead,
		"sales_channel": SalesChannelPostrunRead,
	}

	UpdateFunctions = map[string]func(*request.Request, *model.User, *zap.Logger) error{
		"order": OrderPostrunUpdate,
	}

	ListFunctions = map[string]func(request.GetRequest, any, *model.User) (any, error){
		"user":          UserPostrunList,
		"product":       ProductPostrunList,
		"order":         OrderPostrunList,
		"sales_channel": SalesChannelPostrunList,
	}

	CacheListFunctions = map[string]func(request.GetRequest, request.Response, *model.User, *zap.Logger) error{
		"product":  ProductListCache,
		"category": CategoryListCache,
	}

	CacheCreateFunctions = map[string]func(request.Request, request.Response, *model.User, *zap.Logger) error{
		"product":  ProductCreateCache,
		"category": CategoryCreateCache,
	}

	CacheUpdateFunctions = map[string]func(request.Request, request.Response, *model.User, *zap.Logger) error{
		"product":  ProductUpdateCache,
		"category": CategoryUpdateCache,
	}
)
