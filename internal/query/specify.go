package query

import (
	"bookbox-backend/internal/request"

	"gorm.io/gorm"
)

const (
	defaultLimit  = 10
	maxLimit      = 1001
	defaultOffset = 0
)

func Specify(request request.GetRequest) (orderBy string, err error) {
	//err is unused for now, add if needed

	if request.Metadata.OrderBy.Key != "" &&
		request.Metadata.OrderBy.Type != "" {
		orderBy = request.Metadata.OrderBy.Key + " " + request.Metadata.OrderBy.Type
	}

	return
}

func Paginate(limit, offset int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if offset == 0 || limit == 0 {
			return db
		}

		page := offset
		if page <= 0 {
			page = 1
		}

		pageSize := limit
		switch {
		case pageSize > maxLimit:
			pageSize = maxLimit
		case pageSize <= 0:
			pageSize = defaultLimit
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize + 1)
	}
}
