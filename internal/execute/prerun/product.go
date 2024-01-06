package prerun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"encoding/json"

	"go.uber.org/zap"
)

// ProductPrerunRead prerun functions for user
func ProductPrerunRead(request *request.GetRequest, issuer *model.User) (err error) {
	return
}

// ProductPrerunList prerun functions for user
func ProductPrerunList(listRequest *request.GetRequest, issuer *model.User) (err error) {
	if listRequest.Data.SalesChannelID != "" {
		salesChannel := model.SalesChannel{}
		err = database.DB.Find(&salesChannel, "id = ? OR domain = ?", listRequest.Data.SalesChannelID, listRequest.Data.SalesChannelID).Error
		if err != nil {
			return
		}

		listRequest.Metadata.Relationships = append(listRequest.Metadata.Relationships, request.Relationship{
			Name: "sales_channels",
			RelationParams: []request.FilterParam{
				{
					Key:   "id",
					Value: salesChannel.ID,
					Type:  "eq",
				},
			},
		})
	}

	return
}

// ProductPrerunUpdate prerun functions for user
func ProductPrerunUpdate(request *request.Request, issuer *model.User) (err error) {
	return
}

// ProductPrerunCreate prerun functions for user
func ProductPrerunCreate(request *request.Request, issuer *model.User) (err error) {
	return
}

// ProductPrerunDelete prerun functions for user
func ProductPrerunDelete(request *request.Request, issuer *model.User) (err error) {
	return
}

func ProductPrerunCache(inputRequest request.GetRequest, issuer *model.User, log *zap.Logger) (cached request.Response, found bool, err error) {
	var (
		raw  []byte
		data interface{}
	)

	raw, err = json.Marshal(inputRequest)
	if err != nil {
		return
	}

	data, found = ReadCached(raw, inputRequest)

	if found {
		cached = data.(request.Response)
		log.Info("loaded cache data for request",
			zap.Any("request", inputRequest),
		)
	}

	return
}
