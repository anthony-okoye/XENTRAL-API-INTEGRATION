package crud

import (
	"context"
	"fmt"
	"time"

	"bookbox-backend/internal/database"
	"bookbox-backend/internal/execute/postrun"
	"bookbox-backend/internal/execute/prerun"
	"bookbox-backend/internal/query"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/auth"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ReadHandler(ctx *gin.Context) {

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
		zap.String("entity", readRequest.Entity),
		zap.Any("data", readRequest.Data),
	))

	log.Info("read started")

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

	row := query.Determine(readRequest.Entity)
	if row == nil {
		err = fmt.Errorf("entity does not exist")
		log.Error("Failed to parse input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, readResponse, []string{err.Error()}, 400, log)
		return
	}

	// check authorisation
	if !IsAuthorized(issuer, "read", readRequest.Entity) {
		errMsg := "authorization failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user is not authorized for this request")
		fail.ReturnError(ctx, readResponse, []string{err.Error()}, 403, log)
		return
	}

	// if authorized, add universal filter
	prerun.UniversalFilter(&readRequest, issuer)

	// run prerun functions if they exist
	if f, exist := prerun.ReadFunctions[readRequest.Entity]; exist {
		err = f(&readRequest, issuer)
		if err != nil {
			log.Error("failed to run prerun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, readResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	dbHandler := query.DetermineRelations(readRequest, database.DB)

	res := dbHandler.
		Omit("password").
		Find(row, "id = ?", readRequest.Data.ID)
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

	log.Info("read finished",
		zap.Int64("rowsAffected", res.RowsAffected),
	)

	// run postrun functions if they exist
	if f, exist := postrun.ReadFunctions[readRequest.Entity]; exist {
		row, err = f(readRequest, row, issuer)
		if err != nil {
			log.Error("failed in postrun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, readResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	readResponse.Data = row
	readResponse.Status = true
	ctx.JSON(200, readResponse)
}

const (
	countLabelThreshold = 1000
)

func ListHandler(ctx *gin.Context) {
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

	log := logger.Log.WithOptions(zap.Fields(
		zap.String("entity", listRequest.Entity),
		zap.Any("metadata", listRequest.Metadata),
	))

	log.Info("list started")

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

	rows := query.DetermineArray(listRequest.Entity)
	if rows == nil {
		err = fmt.Errorf("entity does not exist")
		log.Error("Failed to parse input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
		return
	}

	// check authorization
	if !IsAuthorized(issuer, "list", listRequest.Entity) {
		errMsg := "authorization failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user is not authorized for this request")
		fail.ReturnError(ctx, listResponse, []string{err.Error()}, 403, log)
		return
	}

	// run prerun functions if they exist
	if f, exist := prerun.CacheFunctions[listRequest.Entity]; exist {
		data, found, err := f(listRequest, issuer, log)
		if err != nil {
			log.Error("failed to run prerun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
			return
		}

		if found {
			listResponse = data
			listResponse.Status = true
			ctx.JSON(200, listResponse)
			return
		}
	}

	// if authorized, add universal filter
	prerun.UniversalFilter(&listRequest, issuer)

	// run prerun functions if they exist
	if f, exist := prerun.ListFunctions[listRequest.Entity]; exist {
		err = f(&listRequest, issuer)
		if err != nil {
			log.Error("failed to run prerun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
			return
		}
	}

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

	dbContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbHandler := database.DB.WithContext(dbContext)
	dbHandler = query.DetermineRelations(listRequest, dbHandler)

	res := dbHandler.
		Omit("password").
		Scopes(query.Paginate(listRequest.Metadata.Limit, listRequest.Metadata.Offset)).
		Where(where.Main, where.Values...).
		Or(should.Main, should.Values...).
		Order(orderBy).
		Find(rows)

	if res.Error != nil {
		log.Error("list failed",
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, listResponse, []string{fail.SystemError(res.Error)}, 400, log)
		return
	}

	rowsCheck := query.DetermineArray(listRequest.Entity)

	dbHandler = database.DB.WithContext(dbContext)
	dbHandler = query.DetermineRelations(listRequest, dbHandler)

	res = dbHandler.
		Omit("password").
		Scopes(query.Paginate(countLabelThreshold, 1)).
		Where(where.Main, where.Values...).
		Or(should.Main, should.Values...).
		Find(rowsCheck)

	if res.Error != nil {
		log.Error("list failed",
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, listResponse, []string{fail.SystemError(res.Error)}, 400, log)
		return
	}

	totalCheck := query.DetermineCount(listRequest.Entity, rowsCheck)
	listResponse.Total = totalCheck
	if listResponse.Total > 1000 {
		listResponse.Total = 1000
	}

	log.Info("list finished",
		zap.Int64("rowsAffected", res.RowsAffected),
	)

	// run prerun functions if they exist
	if f, exist := postrun.ListFunctions[listRequest.Entity]; exist {
		rows, err = f(listRequest, rows, issuer)
		if err != nil {
			log.Error("failed to run postrun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	total := query.DetermineCount(listRequest.Entity, rows)
	if listRequest.Metadata.Limit != 0 && listRequest.Metadata.Offset != 0 {
		if total > listRequest.Metadata.Limit {
			listResponse.NextOffset = listRequest.Metadata.Offset + 1
			rows = query.CutArrayToLimit(listRequest.Entity, listRequest.Metadata.Limit, rows)
		} else {
			listResponse.NextOffset = -1
		}
	}

	listResponse.Data = rows
	listResponse.Status = true

	// run postrun cache functions if they exist
	if f, exist := postrun.CacheListFunctions[listRequest.Entity]; exist {
		err = f(listRequest, listResponse, issuer, log)
		if err != nil {
			log.Error("failed to run postrun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, listResponse, []string{err.Error()}, 400, log)
			return
		}

		log.Info("saved cache data for request")
	}

	ctx.JSON(200, listResponse)
}

func init() {
	router.Router.Handle("POST", "/read", ReadHandler)
	router.Router.Handle("POST", "/list", ListHandler)
}
