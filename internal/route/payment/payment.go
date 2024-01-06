package payment

import (
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/auth"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"
	"bookbox-backend/pkg/payment"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Initialize(ctx *gin.Context) {
	var (
		paymentResponse = request.Response{}
		paymentRequest  = request.Request{}
	)

	log := logger.Log.WithOptions(zap.Fields())
	log.Info("payment initialize started")

	// check authentication
	_, err := auth.GetIssuer(ctx)
	if err != nil {
		errMsg := "authentication failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, paymentResponse, []string{err.Error()}, 403, log)
		return
	}

	// Parse the posted JSON data.
	err = ctx.ShouldBind(&paymentRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, paymentResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	paymentURL, err := payment.PaymentPageInitialize(paymentRequest)
	if err != nil {
		logger.Log.Error("payment initialize failed",
			zap.Error(err),
		)

		fail.ReturnError(ctx, paymentResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	log.Info("payment initialize finished")

	paymentResponse = request.Response{
		Data: map[string]string{
			"redirectURL": paymentURL,
		},
		Status: true,
	}

	ctx.JSON(200, paymentResponse)
}

func init() {
	router.Router.Handle("POST", "payment/initialize", Initialize)
}
