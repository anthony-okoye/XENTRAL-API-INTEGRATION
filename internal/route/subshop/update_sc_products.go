package subshop

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/auth"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UpdateSCProductsRequest struct {
}

func UpdateSCProductsHandler(ctx *gin.Context) {
	var (
		updateSCProductsRequest = model.SalesChannelProduct{}
		updateResponse          = request.Response{}
	)

	// bind input data to request format
	err := ctx.ShouldBindJSON(&updateSCProductsRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.String("sales_channel_id", updateSCProductsRequest.SalesChannelID),
		zap.String("product_id", updateSCProductsRequest.ProductID),
		zap.Float64("changed_price", updateSCProductsRequest.ChangedPrice),
		zap.String("changed_title", updateSCProductsRequest.ChangedTitle),
	))

	log.Info("update_sc_products started")

	issuer, err := auth.GetIssuer(ctx)
	if err != nil {
		errMsg := "authentication failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 403, log)
		return
	}

	if issuer.Role != "admin" {
		err = fmt.Errorf("only admins can call this route")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	if updateSCProductsRequest.SalesChannelID == "" {
		err = fmt.Errorf("sales channel id is empty")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	if updateSCProductsRequest.ProductID == "" {
		err = fmt.Errorf("product id is empty")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	if updateSCProductsRequest.ChangedTitle == "" {
		err = fmt.Errorf("changed title is empty")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	if updateSCProductsRequest.ChangedPrice == 0 {
		err = fmt.Errorf("changed price is empty")
		log.Error("Data missing fields",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	err = database.DB.Model(&model.SalesChannelProduct{}).Where("product_id = ? AND sales_channel_id = ?", updateSCProductsRequest.ProductID, updateSCProductsRequest.SalesChannelID).Updates(updateSCProductsRequest).Error
	if err != nil {
		log.Error("failed to update",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	log.Info("update_sc_product finished")

	updateResponse.Status = true

	ctx.JSON(200, updateResponse)
}

func init() {
	router.Router.Handle("POST", "/update_sc_products", UpdateSCProductsHandler)
}
