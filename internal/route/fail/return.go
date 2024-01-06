package fail

import (
	"bookbox-backend/internal/request"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Error struct {
	Message string `json:"message"`
	Label   string `json:"label"`
}

func SystemError(requestID any) string {
	return fmt.Sprintf("System error happened. Request ID: %s", requestID)
}

func ReturnError(ctx *gin.Context, response request.Response, errors []string, statusCode int, log *zap.Logger) {
	var (
		raw []byte
		err error
	)

	response.Errors = errors
	response.Status = false

	raw, err = json.Marshal(response)
	if err != nil {
		log.Error("Failed to marshal data",
			zap.Error(err),
		)

		ctx.JSON(statusCode, response)
		return
	}

	ctx.Data(statusCode, "application/json", raw)
}
