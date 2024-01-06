package subshop

import (
	"context"
	"fmt"
	"time"

	"bookbox-backend/internal/database"
	"bookbox-backend/internal/execute/prerun"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/query"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/auth"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ReadSCProductsHandler(ctx *gin.Context) {
	var (
		readRequest  = request.GetRequest{}
		readResponse = request.Response{}
	)

	// bind input data to request format
	err := ctx.ShouldBind(&readRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, readResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.Any("data", readRequest.Data),
	))

	readRequest.Entity = "product"
	if readRequest.Data.SalesChannelID == "" {
		err = fmt.Errorf("sales channel id is empty")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, readResponse, []string{err.Error()}, 400, log)
		return
	}

	if readRequest.Data.ID == "" {
		err = fmt.Errorf("product id is empty")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, readResponse, []string{err.Error()}, 400, log)
		return
	}

	log.Info("read_sc_products started")

	issuer, err := auth.GetIssuer(ctx)
	if err != nil {
		errMsg := "authentication failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, readResponse, []string{err.Error()}, 403, log)
		return
	}

	// if authorized, add universal filter
	prerun.UniversalFilter(&readRequest, issuer)

	dbHandler := query.DetermineRelations(readRequest, database.DB)

	product := &model.Product{}
	res := dbHandler.
		Omit("password").
		Find(product, "id = ?", readRequest.Data.ID)
	if res.Error != nil {
		log.Error("read failed",
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, readResponse, []string{fail.SystemError(res.Error)}, 400, log)
		return
	}

	if res.RowsAffected == 0 {
		log.Warn("no data to read",
			zap.String("id", readRequest.Data.ID),
		)
		return
	}

	scProducts := &model.SalesChannelProduct{}
	res = database.DB.
		Find(scProducts, "product_id = ? and sales_channel_id = ?", readRequest.Data.ID, readRequest.Data.SalesChannelID)
	if res.Error != nil {
		log.Error("read failed",
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, readResponse, []string{fail.SystemError(res.Error)}, 400, log)
		return
	}

	if res.RowsAffected == 0 {
		log.Error("read failed",
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, readResponse, []string{fail.SystemError(res.Error)}, 400, log)
		return
	}

	if scProducts.ChangedPrice != 0 {
		product.SellingPrice = scProducts.ChangedPrice
	}

	if scProducts.ChangedTitle != "" {
		product.Title = scProducts.ChangedTitle
	}

	log.Info("read_sc_products finished",
		zap.Int64("rowsAffected", res.RowsAffected),
	)

	readResponse.Data = product
	readResponse.Status = true
	ctx.JSON(200, readResponse)
}

const (
	countLabelThreshold = 1000
)

func ListSCProductsHandler(ctx *gin.Context) {
	var (
		listRequest  = request.GetRequest{}
		listResponse = request.Response{}
	)

	// bind input data to request format
	err := ctx.ShouldBind(&listRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	listRequest.Entity = "product"
	log := logger.Log.WithOptions(zap.Fields(
		zap.String("entity", listRequest.Entity),
		zap.Any("metadata", listRequest.Metadata),
	))

	if listRequest.Data.SalesChannelID == "" {
		err = fmt.Errorf("sales channel id is empty")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
		return
	}

	listRequest.Metadata.Relationships = append(listRequest.Metadata.Relationships, request.Relationship{
		Name: "sales_channels",
		RelationParams: []request.FilterParam{
			{
				Key:   "id",
				Value: listRequest.Data.SalesChannelID,
				Type:  "eq",
			},
		},
	})

	log.Info("list_sc_products started")

	issuer, err := auth.GetIssuer(ctx)
	if err != nil {
		errMsg := "authentication failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, listResponse, []string{err.Error()}, 403, log)
		return
	}

	// if authorized, add universal filter
	prerun.UniversalFilter(&listRequest, issuer)

	where, should, err := query.Constraints(listRequest)
	if err != nil {
		log.Error("Failed to parse input filters",
			zap.Error(err),
			zap.Any("filter", listRequest.Metadata.Filter),
		)

		fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
		return
	}

	orderBy, err := query.Specify(listRequest)
	if err != nil {
		log.Error("Failed to parse input specification",
			zap.Error(err),
			zap.Any("orderBy", listRequest.Metadata.OrderBy),
		)

		fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
		return
	}

	dbContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbHandler := database.DB.WithContext(dbContext)
	dbHandler = query.DetermineRelations(listRequest, dbHandler)

	rows := []model.Product{}

	res := dbHandler.
		Omit("password").
		Scopes(query.Paginate(listRequest.Metadata.Limit, listRequest.Metadata.Offset)).
		Where(where.Main, where.Values...).
		Or(should.Main, should.Values...).
		Order(orderBy).
		Find(&rows)

	if res.Error != nil {
		log.Error("list_sc_products failed",
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, listResponse, []string{fail.SystemError(res.Error)}, 400, log)
		return
	}

	rowsCheck := []model.Product{}

	dbHandler = database.DB.WithContext(dbContext)
	dbHandler = query.DetermineRelations(listRequest, dbHandler)

	res = dbHandler.
		Omit("password").
		Scopes(query.Paginate(countLabelThreshold, 1)).
		Where(where.Main, where.Values...).
		Or(should.Main, should.Values...).
		Find(&rowsCheck)

	if res.Error != nil {
		log.Error("list_sc_products failed",
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, listResponse, []string{fail.SystemError(res.Error)}, 400, log)
		return
	}

	totalCheck := len(rowsCheck)
	listResponse.Total = totalCheck
	if listResponse.Total > 1000 {
		listResponse.Total = 1000
	}

	total := len(rows)
	if listRequest.Metadata.Limit != 0 && listRequest.Metadata.Offset != 0 {
		if total > listRequest.Metadata.Limit {
			listResponse.NextOffset = listRequest.Metadata.Offset + 1
			rows = rows[0:listRequest.Metadata.Limit]
		} else {
			listResponse.NextOffset = -1
		}
	}

	for i, row := range rows {
		scProducts := &model.SalesChannelProduct{}
		res = database.DB.
			Find(scProducts, "product_id = ? and sales_channel_id = ?", row.ID, listRequest.Data.SalesChannelID)
		if res.Error != nil {
			log.Error("read failed",
				zap.Error(res.Error),
			)

			fail.ReturnError(ctx, listResponse, []string{fail.SystemError(res.Error)}, 400, log)
			return
		}

		if res.RowsAffected == 0 {
			log.Error("read failed",
				zap.Error(res.Error),
			)

			fail.ReturnError(ctx, listResponse, []string{fail.SystemError(res.Error)}, 400, log)
			return
		}

		if scProducts.ChangedPrice != 0 {
			rows[i].SellingPrice = scProducts.ChangedPrice
		}

		if scProducts.ChangedTitle != "" {
			rows[i].Title = scProducts.ChangedTitle
		}
	}

	log.Info("list_sc_products finished",
		zap.Int64("rowsAffected", res.RowsAffected),
	)

	listResponse.Data = rows
	listResponse.Status = true

	ctx.JSON(200, listResponse)
}

func init() {
	router.Router.Handle("POST", "/read_sc_products", ReadSCProductsHandler)
	router.Router.Handle("POST", "/list_sc_products", ListSCProductsHandler)
}
