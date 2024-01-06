package processor

import (
	"bookbox-backend/pkg/logger"
)

const (
	DefaultRetryCount = 25
)

func init() {
	log := logger.Log

	go Process(log)
}
