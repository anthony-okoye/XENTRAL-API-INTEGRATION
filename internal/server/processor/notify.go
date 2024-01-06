package processor

import (
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/server/sendgrid"
	"fmt"
	"os"

	"go.uber.org/zap"
)

var (
	senderEmail, orderTemplateId, failedOrderTemplateId string
)

func init() {
	senderEmail = os.Getenv("SENDGRID_SENDER_EMAIL")
	orderTemplateId = os.Getenv("SENDGRID_ORDER_TEMPLATE_ID")
	failedOrderTemplateId = os.Getenv("SENDGRID_FAILED_TEMPLATE_ID")
}

func SendOrderNotification(order *model.OrderNotification, log *zap.Logger) (err error) {
	var (
		notify sendgrid.SendGrid
	)

	orderDetails := make([]map[string]string, 0)

	for _, item := range order.Details.Products {
		orderDetails = append(orderDetails, map[string]string{
			"title":     item.Product.Title,
			"subtitle":  item.Product.Subtitle,
			"ebook_url": item.EBookURL,
			"price":     fmt.Sprintf("%.2f€", item.Product.SellingPrice),
		})
	}

	// add the personalization object and set the dynamic template data
	notify.DynamicTemplateData = map[string]interface{}{
		"address":     order.Details.DeliveryAddress,
		"order_id":    order.Details.ID,
		"order":       orderDetails,
		"total_price": fmt.Sprintf("%.2f€", order.Details.TotalPrice),
	}

	notify.From = senderEmail
	notify.To = order.Details.Email
	notify.TemplateID = orderTemplateId

	err = notify.SendEmail()
	if err != nil {
		log.Error("failed to send email",
			zap.Error(err),
		)
		err = nil
	}

	return
}

func SendOrderFailedNotification(order *model.OrderNotification, log *zap.Logger) (err error) {
	var (
		notify sendgrid.SendGrid
	)

	notify.DynamicTemplateData = map[string]interface{}{
		"order_id": order.Details.ID,
	}
	notify.From = senderEmail
	notify.To = order.Details.Email
	notify.TemplateID = failedOrderTemplateId

	err = notify.SendEmail()
	if err != nil {
		log.Error("failed to send email",
			zap.Error(err),
		)
		err = nil
	}

	return
}
