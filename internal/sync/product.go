package sync

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

var failed = 0

func batchUpdate(products []model.Product, wg *sync.WaitGroup) (err error) {
	for i := 0; i < len(products); i++ {
		// Start a transaction
		tx := database.DB.Begin()

		if len(products[i].Categories) != 0 {
			if err := tx.Exec("DELETE FROM product_categories WHERE product_id = ?", products[i].ID).Error; err != nil {
				tx.Rollback()
				logger.Log.Error("failed to update",
					zap.Error(err),
				)
				failed++
				continue
			}
		}

		if err := tx.Updates(&products[i]).Error; err != nil {
			// Handle the error
			tx.Rollback()

			logger.Log.Error("failed to update",
				zap.Error(err),
			)
			failed++
			continue
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			// Handle the error
			tx.Rollback()
			logger.Log.Error("failed to update",
				zap.Error(err),
			)
			failed++
			continue
		}

		if i%1000 == 0 {
			time.Sleep(time.Millisecond * 50)
		}
	}

	wg.Done()

	return nil
}

func batchCreate(products []model.Product, batchSize int) (err error) {
	err = database.DB.CreateInBatches(products, batchSize).Error
	if err != nil {
		return err
	}

	return
}

func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}
