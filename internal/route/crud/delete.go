package crud

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/execute/prerun"
	"bookbox-backend/internal/query"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/auth"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func DeleteHandler(ctx *gin.Context) {
	var (
		deleteRequest  = request.Request{}
		deleteResponse = request.Response{}
	)

	// bind input data to request format
	err := ctx.ShouldBindJSON(&deleteRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.String("entity", deleteRequest.Entity),
		zap.Any("data", deleteRequest.Data),
	))

	log.Info("delete started")

	issuer, err := auth.GetIssuer(ctx)
	if err != nil {
		errMsg := "authentication failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 403, log)
		return
	}

	if deleteRequest.Entity == "product" {
		productMutex.Lock()
		defer productMutex.Unlock()
	}

	row := query.Determine(deleteRequest.Entity)
	raw, err := json.Marshal(deleteRequest.Data)
	if err != nil {
		log.Error("Failed to marshal data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 400, log)
		return
	}

	// check authorisation
	if !IsAuthorized(issuer, "delete", deleteRequest.Entity) {
		errMsg := "authorization failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user is not authorized for this request")
		fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 403, log)
		return
	}

	// run prerun functions if they exist
	if f, exist := prerun.DeleteFunctions[deleteRequest.Entity]; exist {
		err = f(&deleteRequest, issuer)
		if err != nil {
			log.Error("failed to run prerun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	id := deleteRequest.Data["id"].(string)

	err = json.Unmarshal(raw, &row)
	if err != nil {
		log.Error("Failed to unmarshal data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 400, log)
		return
	}

	res := database.DB.Delete(&row, "id = ?", id)
	if res.Error != nil {
		err := fmt.Errorf("failed to update rows, wrong id or product already removed")
		log.Error("Delete failed",
			zap.String("id", id),
			zap.Error(res.Error),
		)

		fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 400, log)
		return
	}

	if res.RowsAffected == 0 {
		err := fmt.Errorf("failed to update rows, wrong id or product already removed")
		log.Error("Delete failed",
			zap.String("id", id),
			zap.Error(err),
		)

		fail.ReturnError(ctx, deleteResponse, []string{err.Error()}, 400, log)
		return
	}

	log.Info("delete finished",
		zap.Any("data", deleteRequest.Data),
	)

	deleteResponse.Status = true
	ctx.JSON(200, deleteResponse)
}

func init() {
	router.Router.Handle("POST", "/delete", DeleteHandler)
}
