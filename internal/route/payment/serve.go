package payment

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/execute/postrun"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/crud"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/processor"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"
	"bookbox-backend/pkg/payment"
	"bookbox-backend/pkg/redis"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	b64 "encoding/base64"
	"encoding/json"
)

func ServeStatic(ctx *gin.Context) {
	var (
		serveResponse request.Response
	)

	log := logger.Log.WithOptions(zap.Fields())

	log.Info("payment serve started")

	// Check if params template is being served
	data := ctx.Query("data")
	if data == "" {
		log.Error("query param 'data' is empty")

		// return status:false to ST
		fail.ReturnError(ctx, serveResponse, []string{"query param 'data' is empty"}, 400, log)
		return
	}

	decoded, err := b64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Error("decoding failed",
			zap.String("data", data),
			zap.String("decoded", string(decoded)),
			zap.Error(err),
		)

		// return status:false to ST
		fail.ReturnError(ctx, serveResponse, []string{"decoding failed"}, 400, log)
		return
	}

	requestID := string(decoded)
	storedJSON, err := redis.Client.Get(requestID).Result()
	if err != nil {
		log.Error("failed to get request data from Redis",
			zap.Error(err),
		)

		fail.ReturnError(ctx, serveResponse, []string{"failed to get request data"}, 400, log)
		return
	}

	stored := payment.PaymentStorage{}
	err = json.Unmarshal([]byte(storedJSON), &stored)
	if err != nil {
		log.Error("failed to unmarshal data",
			zap.Error(err),
		)

		ctx.Redirect(302, stored.ErrorReturnURL)
		return
	}

	row := model.Order{}
	res := database.DB.Find(&row, "id = ?", stored.OrderID)
	if res.Error != nil {
		log.Error("failed to read order from DB",
			zap.Error(err),
		)

		ctx.Redirect(302, stored.ErrorReturnURL)
		return
	}

	if res.RowsAffected == 0 {
		log.Error("failed to read order from DB")

		ctx.Redirect(302, stored.ErrorReturnURL)
		return
	}

	if !(*row.Active) {
		log.Error("order with given id is not available",
			zap.Error(err),
		)

		ctx.Redirect(302, stored.ErrorReturnURL)
		return
	}

	log = logger.Log.WithOptions(zap.Fields(
		zap.String("orderId", stored.OrderID),
	))

	isPaid, err := payment.PaymentAuthorize(stored, log)
	if err != nil {
		log.Error("error while authorizing",
			zap.Error(err),
		)

		row.PaymentStatus = "failed"
		err = crud.UpdateTransaction(request.Request{}, database.DB, row, stored.OrderID)
		if err != nil {
			log.Error("failed to update payment status",
				zap.Error(err),
			)

			ctx.Redirect(302, stored.ErrorReturnURL)
			return
		}

		ctx.Redirect(302, stored.ErrorReturnURL)
		return
	}

	if !isPaid {
		err = fmt.Errorf("failed to verify")
		log.Error("not paid",
			zap.Error(err),
		)

		row.PaymentStatus = "failed"
		err = crud.UpdateTransaction(request.Request{}, database.DB, row, stored.OrderID)
		if err != nil {
			log.Error("failed to update payment status",
				zap.Error(err),
			)

			ctx.Redirect(302, stored.ErrorReturnURL)
			return
		}

		ctx.Redirect(302, stored.ErrorReturnURL)
		return
	}

	// update payment status to paid in database
	row.PaymentStatus = "paid"
	err = crud.UpdateTransaction(request.Request{}, database.DB, row, stored.OrderID)
	if err != nil {
		log.Error("failed to update payment status",
			zap.Error(err),
		)

		ctx.Redirect(302, stored.ErrorReturnURL)
		return
	}

	id := row.ID
	err = processor.ProcessOrder(id, logger.Log)
	if err != nil {
		ctx.Redirect(302, stored.ReturnURL)
		return
	}

	err = postrun.ProcessPaidOrder(row)
	if err != nil {
		ctx.Redirect(302, stored.ReturnURL)
		return
	}
	log.Info("xentral order processing starting")

	log.Info("payment serve finished")

	ctx.Redirect(302, stored.ReturnURL)
}

func init() {
	router.Router.Handle("GET", "payment/verify", ServeStatic)
}
