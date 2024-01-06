package prerun

import (
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
)

func UniversalFilter(req *request.GetRequest, issuer *model.User) (err error) {

	universalFilter := request.FilterParam{
		Key:   "active",
		Value: "true",
		Type:  "eq",
	}

	if issuer.Role != "admin" {
		req.Metadata.Filter.Must = append(req.Metadata.Filter.Must, universalFilter)
	}

	if issuer.Role == "admin" && req.Metadata.OrderBy.Key == "" {
		req.Metadata.OrderBy = request.OrderBy{
			Key:  "updated_at",
			Type: "desc",
		}
	}

	return
}
