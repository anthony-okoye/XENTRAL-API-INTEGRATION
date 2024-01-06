package processor

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/ebooks"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

func Process(log *zap.Logger) {
	log.Info("ebooks order processor started")

	for {
		time.Sleep(time.Second * 5)

		files, err := os.ReadDir(ebooks.QueuePath)
		if err != nil {
			continue
		}

		// process the orders
		for i := 0; i < len(files); i++ {
			filePath := filepath.Join(ebooks.QueuePath, files[i].Name())
			order, err := ebooks.ReadOrder(filePath)
			if err != nil {
				log.Debug("failed to read file",
					zap.Error(err),
					zap.String("filepath", filePath),
				)
				continue
			}

			if order.RetryCount <= 0 {
				log.Debug("moving order to dead letter queue",
					zap.String("orderId", order.Details.ID),
					zap.String("filepath", filePath),
				)

				// move the order to dead letter queue
				err = ebooks.SaveOrder(ebooks.DeadLetterQueuePath, order)
				if err != nil {
					log.Error("moving order to dead letter queue failed",
						zap.String("orderId", order.Details.ID),
						zap.Error(err),
					)
					return
				}

				log.Debug("sending failed order notification",
					zap.String("filepath", filePath),
					zap.String("orderId", order.Details.ID),
				)

				// notify the user about the failed order notification
				err = SendOrderFailedNotification(&order, log)
				if err != nil {
					log.Error("send failed order notification failed",
						zap.String("id", order.Details.ID),
						zap.Error(err),
					)
				}

				// remove order from orders to do
				os.Remove(filePath)
				continue
			}

			order.RetryCount--
			ebooks.SaveOrder(ebooks.QueuePath, order)

			err = ebooks.GetDownloadUrl(&order, order.EbooksCount)
			if err != nil {
				log.Error("failed to get urls",
					zap.String("orderId", order.Details.ID),
					zap.Error(err),
				)

				continue
			}

			log.Debug("sending order notification",
				zap.Any("items", order.Details.Products),
			)

			err = SendOrderNotification(&order, log)
			if err != nil {
				log.Error("send order notification failed",
					zap.String("id", order.Details.ID),
					zap.Error(err),
				)

				continue
			}

			log.Debug("send order notification successful",
				zap.String("id", order.Details.ID),
			)

			// remove order from orders to do
			err = os.Remove(filePath)
			if err != nil {
				log.Error("failed to remove file",
					zap.String("filepath", filePath),
					zap.Error(err),
				)
				continue
			}
		}
	}
}

func ProcessOrder(id string, log *zap.Logger) (err error) {
	order := model.Order{}

	res := database.DB.Preload("Products.Product").Preload("User").Where("id = ?", id).First(&order)
	if res.Error != nil {
		errMsg := "failed to read order in postrun"
		log.Error(errMsg,
			zap.Error(err),
		)
		return
	}

	for i := range order.Products {
		order.Products[i].Product.SellingPrice = order.Products[i].CurrentPrice
	}

	saveOrder := model.OrderNotification{
		Details:    order,
		RetryCount: 25,
	}

	eBookIds := make([]string, 0)

	for i := 0; i < len(order.Products); i++ {
		// if the product has eBook option, add it's ID to the pool
		if order.Products[i].Product.IsDownloadTitle {
			eBookIds = append(eBookIds, order.Products[i].Product.ID)
		}
	}

	if len(eBookIds) != 0 {
		// create eBook order with all id's at once
		saveOrder.Details.EBookOrderID, err = ebooks.CreateOrder(eBookIds)
		if err != nil {
			log.Info("sending failed order notification",
				zap.Any("items", saveOrder.Details.Products),
			)

			err = SendOrderFailedNotification(&saveOrder, log)
			if err != nil {
				log.Error("send failed order notification failed",
					zap.String("id", saveOrder.Details.ID),
					zap.Error(err),
				)
			}

			log.Info("send failed order notification successful",
				zap.String("id", saveOrder.Details.ID),
			)

			err = ebooks.SaveOrder(ebooks.DeadLetterQueuePath, saveOrder)
			if err != nil {
				log.Error("moving order to dead letter queue failed",
					zap.String("orderId", saveOrder.Details.ID),
					zap.Error(err),
				)
				return
			}

			return
		}

		saveOrder.EbooksCount = len(eBookIds)
	}

	err = ebooks.SaveOrder(ebooks.QueuePath, saveOrder)
	if err != nil {
		return
	}

	return
}
