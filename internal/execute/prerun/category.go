package prerun

import (
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"encoding/json"

	"go.uber.org/zap"
)

func CategoryPrerunCache(inputRequest request.GetRequest, issuer *model.User, log *zap.Logger) (cached request.Response, found bool, err error) {
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
