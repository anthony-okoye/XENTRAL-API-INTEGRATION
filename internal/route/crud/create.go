package crud

import (
	"encoding/json"
	"fmt"

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

func CreateHandler(ctx *gin.Context) {
	var (
		createRequest  = request.Request{}
		createResponse = request.Response{}
	)

	// bind input data to request format
	err := ctx.ShouldBind(&createRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.String("entity", createRequest.Entity),
	))

	log.Info("create started")

	// check authentication
	issuer, err := auth.GetIssuer(ctx)
	if err != nil {
		errMsg := "authentication failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, createResponse, []string{err.Error()}, 403, log)
		return
	}

	if createRequest.Entity == "order" {
		productMutex.Lock()
		defer productMutex.Unlock()
	}

	// determines entity
	row := query.Determine(createRequest.Entity)
	if row == nil {
		err = fmt.Errorf("entity does not exist")
		log.Error("Failed to parse input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, log)
		return
	}

	// check authorisation
	if !IsAuthorized(issuer, "create", createRequest.Entity) {
		errMsg := "authorization failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user is not authorized for this request")
		fail.ReturnError(ctx, createResponse, []string{err.Error()}, 403, log)
		return
	}

	// run prerun functions if they exist
	if f, exist := prerun.CreateFunctions[createRequest.Entity]; exist {
		err = f(&createRequest, issuer)
		if err != nil {
			log.Error("failed in prerun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	raw, err := json.Marshal(createRequest.Data)
	if err != nil {
		log.Error("Failed to marshal data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, log)
		return
	}

	err = json.Unmarshal(raw, row)
	if err != nil {
		log.Error("Failed to unmarshal data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, log)
		return
	}

	err = database.DB.Create(row).Error
	if err != nil {
		log.Error("Failed to create data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, log)
		return
	}

	log.Info("create finished")

	// run postrun functions if they exist
	if f, exist := postrun.CreateFunctions[createRequest.Entity]; exist {
		row, err = f(row, issuer)
		if err != nil {
			log.Error("failed in postrun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	response := make(map[string]any)
	response["id"] = query.GetID(row, createRequest.Entity)

	createResponse.Data = response
	createResponse.Status = true

	// run postrun cache functions if they exist
	if f, exist := postrun.CacheCreateFunctions[createRequest.Entity]; exist {
		err = f(createRequest, createResponse, issuer, log)
		if err != nil {
			log.Error("failed to run postrun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, createResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	ctx.JSON(200, createResponse)
}

func init() {
	router.Router.Handle("POST", "/create", CreateHandler)
}
