package sendgrid

import (
	"bookbox-backend/pkg/logger"
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

var (
	apiKey string
)

type SendGrid struct {
	From                string                 `json:"from"`
	To                  string                 `json:"to"`
	DynamicTemplateData map[string]interface{} `json:"dynamicTemplateData"`
	TemplateID          string                 `json:"templateId"`
}

func init() {
	apiKey = os.Getenv("SENDGRID_API_KEY_DEV")
}

func (data *SendGrid) SendEmail() (err error) {
	from := mail.NewEmail("Bookbox", data.From)
	to := mail.NewEmail("Recipient Name", data.To)

	message := mail.NewV3Mail()
	message.SetFrom(from)

	p := mail.NewPersonalization()
	p.DynamicTemplateData = data.DynamicTemplateData
	p.AddTos(to)
	message.AddPersonalizations(p)

	// specify the template ID
	message.SetTemplateID(data.TemplateID)

	// send email

	logger.Log.Debug("sending main",
		zap.String("from", data.From),
		zap.String("to", data.To),
	)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode > 300 {
		err = fmt.Errorf("response status code: %d, response: %s", response.StatusCode, response.Body)
		return
	}

	return nil
}
