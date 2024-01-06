package postrun

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/server/processor"
	"bookbox-backend/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

func OrderPostrunRead(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	order := queried.(*model.Order)
	defer func() {
		response = order
	}()

	for i, val := range order.Products {
		scProducts := &model.SalesChannelProduct{}
		res := database.DB.
			Find(scProducts, "product_id = ? and sales_channel_id = ?", val.ProductID, order.SalesChannelID)
		if res.Error != nil {
			return
		}

		if res.RowsAffected == 0 {
			return
		}

		if scProducts.ChangedPrice != 0 {
			order.Products[i].Product.SellingPrice = scProducts.ChangedPrice
		}

		if scProducts.ChangedTitle != "" {
			order.Products[i].Product.Title = scProducts.ChangedTitle
		}
	}

	return
}

func OrderPostrunList(request request.GetRequest, queried any, issuer *model.User) (response any, err error) {
	orders := queried.(*[]model.Order)
	defer func() {
		response = orders
	}()

	for i := range *orders {
		for j, val := range (*orders)[i].Products {
			scProducts := &model.SalesChannelProduct{}
			res := database.DB.
				Find(scProducts, "product_id = ? and sales_channel_id = ?", val.ProductID, (*orders)[i].SalesChannelID)
			if res.Error != nil {
				return
			}

			if res.RowsAffected == 0 {
				return
			}

			if scProducts.ChangedPrice != 0 {
				(*orders)[i].Products[j].Product.SellingPrice = scProducts.ChangedPrice
			}

			if scProducts.ChangedTitle != "" {
				(*orders)[i].Products[j].Product.Title = scProducts.ChangedTitle
			}
		}
	}

	return
}

func OrderPostrunCreate(queried any, issuer *model.User) (response any, err error) {
	tx := database.DB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	order := model.Order{}
	raw, err := json.Marshal(queried)
	if err != nil {
		return
	}

	json.Unmarshal(raw, &order)

	rawOrder, err := json.Marshal(order)
	if err != nil {
		logger.Log.Error("Failed to marshal order data", zap.Error(err))
		// Handle the error if needed
	} else {
		logger.Log.Info("Order Details:", zap.String("order", string(rawOrder)))
	}

	for _, orderItem := range order.Products {
		row := model.Product{}
		db := database.DB.Where("id = ?", orderItem.ProductID).First(&row)
		if db.Error != nil {
			err = tx.Error
			return
		}

		if db.RowsAffected == 0 {
			err = fmt.Errorf("product does not exist")
			return
		}

		var count int64
		db.Count(&count)
		if count == 0 {
			return nil, fmt.Errorf("product with specified id does not exist")
		}

		// decrease order -> products stock, for future orders
		row.Stock -= orderItem.Quantity

		if row.Stock < 0 {
			err = fmt.Errorf("product: %s is out of stock", row.Title)
			return
		}

		// add updated version
		err = tx.Updates(&row).Error
		if err != nil {
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		return
	}

	response = queried

	if order.PaymentStatus == "paid" {
		id := order.ID

		err = ProcessPaidOrder(order)
		if err != nil {
			logger.Log.Error("Failed to run ProcessPaidOrder func", zap.Error(err))
			return
		}
		logger.Log.Info("xentral order processing starting")

		err = processor.ProcessOrder(id, logger.Log)
		if err != nil {
			return
		}

	}

	return
}

// if order is paid, send email
func OrderPostrunUpdate(request *request.Request, issuer *model.User, log *zap.Logger) (err error) {
	paymentStatus, _ := request.Data["payment_status"].(string)

	if paymentStatus == "paid" {
		id := request.Data["id"].(string)
		err = processor.ProcessOrder(id, log)
		if err != nil {
			return err
		}
	}

	return
}

func ProcessPaidOrder(order model.Order) error {
	var err error
	// Check if user details are available
	if order.UserID != nil && *order.UserID != "" {
		logger.Log.Info("Processing user details started FIRST FUNC")
		// Retrieve user details from the database based on user_id
		user := &model.User{}
		userQuery := database.DB.Where("id = ?", *order.UserID).First(user)
		if userQuery.Error != nil {
			return userQuery.Error
		}

		// Check if user has a delivery address
		if user.DeliveryAddressID != nil {
			// Retrieve address details based on delivery_address_id
			address := &model.Address{}
			addressQuery := database.DB.Where("id = ?", *user.DeliveryAddressID).First(address)
			if addressQuery.Error != nil {
				return addressQuery.Error
			}

			// Pass the order, user, and address details to the function
			err = ProcessUserDetails(&order, user, address, logger.Log)
			if err != nil {
				logger.Log.Error("Failed to make POST request for user details", zap.Error(err))
				// Log an error, but continue with the processing
			}
		}

	}

	// Check if product details are empty
	if len(order.Products) == 0 {
		logger.Log.Error("Product details are empty or null")
		// Log an error, but continue with the processing
	}

	logger.Log.Info("OrderPostrunCreate: Started processing order")

	for _, orderItem := range order.Products {
		// Check if product_id is present in the order item
		if orderItem.ProductID == "" {
			continue
		}

		// Retrieve product details using product_id
		productDetails, err := getProductDetails(orderItem.ProductID)
		if err != nil {
			logger.Log.Error("Failed to retrieve product details", zap.Error(err))
			return err
		}
		logger.Log.Info("OrderPostrunCreate: Retrieved product details", zap.String("product_id", orderItem.ProductID))

		// Use the product details to construct the payload for the API request
		payload, err := constructProductPayload(productDetails)
		if err != nil {
			// Handle the error, such as logging or returning an error response
			logger.Log.Error("Failed to construct product payload", zap.Error(err))
			return err
		}
		logger.Log.Info("OrderPostrunCreate: Constructed product payload", zap.String("payload", string(payload)))

		// Make the API request to create the product
		err = createProductAPIRequest(payload)
		if err != nil {
			logger.Log.Error("Failed to make API request to create product", zap.Error(err))
			return err
		}
		logger.Log.Info("OrderPostrunCreate: Created product successfully", zap.String("product_id", orderItem.ProductID))

	}

	// Process sales order details
	err = ProcessSalesOrderDetails(order, logger.Log)
	if err != nil {
		logger.Log.Error("Failed to make POST request for sales order details", zap.Error(err))
		// Log an error, but continue with the processing
	}
	logger.Log.Info("OrderPostrunCreate: Completed processing product order")

	return nil
}

// Helper function to retrieve product details using product_id
func getProductDetails(productID string) (*model.Product, error) {
	// Query the database or make an API request to retrieve product details
	product := &model.Product{}
	db := database.DB.Where("id = ?", productID).First(product)
	if db.Error != nil {
		return nil, db.Error
	}

	if db.RowsAffected == 0 {
		return nil, fmt.Errorf("product with specified id does not exist")
	}

	return product, nil
}

// Helper function to construct the payload for the API request
func constructProductPayload(productDetails *model.Product) ([]byte, error) {
	// Check if measurements are available, use default values otherwise
	width, height, length, weight := 1.0, 1.0, 1.0, 1.0
	// Convert measurements from millimeters to centimeters and grams to kilograms
	if productDetails.Width != "" {
		widthMM := convertStringToFloat(productDetails.Width)
		width = widthMM / 10.0 // converting millimeters to centimeters
	}
	if productDetails.Height != "" {
		heightMM := convertStringToFloat(productDetails.Height)
		height = heightMM / 10.0 // converting millimeters to centimeters
	}
	if productDetails.Length != "" {
		lengthMM := convertStringToFloat(productDetails.Length)
		length = lengthMM / 10.0 // converting millimeters to centimeters
	}
	if productDetails.Weight != "" {
		weightGrams := convertStringToFloat(productDetails.Weight)
		weight = weightGrams / 1000.0 // converting grams to kilograms
	}

	// Check if description is available, use it; otherwise, set to null
	var description string
	if productDetails.Description != "" {
		description = productDetails.Description
	} else {
		description = "no value"
	}

	// Check if manufacturer name is available, use it; otherwise, set to null
	var publisher string
	if productDetails.Publisher != "" {
		publisher = productDetails.Publisher
	} else {
		publisher = "no value"
	}

	// Check if product name is available, use it; otherwise, set to "no value"
	var productName string
	if productDetails.Title != "" {
		productName = productDetails.Title
	} else {
		productName = "no value"
	}

	// Construct the payload using a map
	payloadMap := map[string]interface{}{
		"project": map[string]interface{}{"id": "1"},
		"measurements": map[string]interface{}{
			"width": map[string]interface{}{
				"unit":  "cm",
				"value": json.Number(fmt.Sprintf("%.2f", width)),
			},
			"height": map[string]interface{}{
				"unit":  "cm",
				"value": json.Number(fmt.Sprintf("%.2f", height)),
			},
			"length": map[string]interface{}{
				"unit":  "cm",
				"value": json.Number(fmt.Sprintf("%.2f", length)),
			},
			"weight": map[string]interface{}{
				"unit":  "kg",
				"value": json.Number(fmt.Sprintf("%.2f", weight)),
			},
			"netWeight": map[string]interface{}{
				"unit":  "kg",
				"value": json.Number(fmt.Sprintf("%.2f", weight)),
			},
		},
		"name":             productName,
		"number":           productDetails.ID,
		"description":      description,
		"ean":              productDetails.EAN,
		"shopPriceDisplay": fmt.Sprintf("%.2f", productDetails.SellingPrice),
		"manufacturer": map[string]interface{}{
			"name":   publisher,
			"number": "no value",
			"link":   "https://no_value",
		},
		"isStockItem":          true,
		"minimumOrderQuantity": 1,
	}

	// Convert the map to JSON
	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		// Handle the error if needed
		return nil, err
	}

	return payloadJSON, nil
}

// Helper function to convert string to float
func convertStringToFloat(str string) float64 {
	floatValue, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return floatValue
}

// Helper function to make the API request to create the product
func createProductAPIRequest(payload []byte) error {
	url := "https://652fcc41115ea.xentral.biz/api/products"

	req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.Header.Add("accept", "text/html")
	req.Header.Add("content-type", "application/vnd.xentral.default.v1+json")
	req.Header.Add("authorization", "Bearer 9|X63RhGycfCWFIV8FmNZw8YliZ11Vcs9Np99k9VT8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	logger.Log.Info("Response from create product API:", zap.String("response", string(body)))

	return nil
}

func ProcessUserDetails(order *model.Order, user *model.User, address *model.Address, log *zap.Logger) error {
	log.Info("Processing user details started")

	url := "https://652fcc41115ea.xentral.biz/api/customers"

	// Step 1: Check if the user already exists
	customerID, err := getCustomerID(order.Email, log)
	if err != nil {
		log.Error("Failed to get customer ID", zap.Error(err))
		return err
	}

	// Step 2: If the user exists, skip user details processing and move to processing product
	/*if customerID != "" {
		log.Info("User already exists. Skipping user details processing.")
		return nil
	}*/

	if customerID == "not found" {
		log.Info("User not found. Proceeding with user details processing.")

	} else if customerID != "" {
		log.Info("User already exists. Skipping user details processing.")
	}

	// Step 3: User does not exist, proceed with processing user details

	// Check the value of user's salutation
	var salutationType string
	switch user.Salutation {
	case "Herr":
		salutationType = "mr"
	case "Frau":
		salutationType = "ms"
	default:
		salutationType = "company" // Use a default value or handle as needed
	}

	// Assuming order.Contact.Email is the user's email field in the order
	payload := fmt.Sprintf(`{
        "type": "%s",
        "general": {
            "name": "%s %s",
            "address": {
                "street": "%s",
                "zip": "%s",
                "city": "%s",
                "state": "%s",
                "country": "%s",
                "note": "User data"
              }
        },
        "contact": {
            "email": "%s",
            "marketingMails": false,
            "trackingMails": false
        }
    }`, salutationType, order.FirstName, order.LastName, address.Street, address.ZipCode, address.City, address.City, address.Country, order.Email)

	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
	req.Header.Add("accept", "text/html")
	req.Header.Add("content-type", "application/vnd.xentral.default.v1+json")
	req.Header.Add("authorization", "Bearer 9|X63RhGycfCWFIV8FmNZw8YliZ11Vcs9Np99k9VT8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	log.Info("Response from user details API:", zap.String("response", string(body)))

	return nil
}

// Helper function to get customer ID
func getCustomerID(email string, log *zap.Logger) (string, error) {
	url := fmt.Sprintf("https://652fcc41115ea.xentral.biz/api/customers?filter[0][key]=email&filter[0][value]=%s&filter[0][op]=equals", email)
	response, err := makeGETRequest(url, log)
	if err != nil {
		return "", err
	}

	log.Info("Response from getCustomerID API:", zap.ByteString("response", response))

	var customerData map[string]interface{}
	err = json.Unmarshal(response, &customerData)
	if err != nil {
		return "", err
	}

	customerDataSlice, ok := customerData["data"].([]interface{})
	if !ok {
		// Handle the case where the data field is not a slice
		return "", fmt.Errorf("customer data is not a slice")
	}

	if len(customerDataSlice) == 0 {
		// Handle the case where the slice is empty
		return "not found", nil
	}

	// Assuming the first customer in the list
	customerID := customerDataSlice[0].(map[string]interface{})["id"].(string)
	return customerID, nil
}

// Helper function to get product ID
func getProductID(productID string, log *zap.Logger) (string, error) {
	url := fmt.Sprintf("https://652fcc41115ea.xentral.biz/api/products?filter[0][key]=number&filter[0][value]=%s&filter[0][op]=equals", productID)
	response, err := makeGETRequest(url, log)
	if err != nil {
		return "", err
	}

	var productData map[string]interface{}
	err = json.Unmarshal(response, &productData)
	if err != nil {
		return "", err
	}

	// Assuming the first product in the list
	productIDValue := productData["data"].([]interface{})[0].(map[string]interface{})["id"].(string)
	return productIDValue, nil
}

// Helper function to make a GET request
func makeGETRequest(url string, log *zap.Logger) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/vnd.xentral.default.v1+json")
	req.Header.Add("authorization", "Bearer 9|X63RhGycfCWFIV8FmNZw8YliZ11Vcs9Np99k9VT8")

	res, err := http.DefaultClient.Do(req)

	if res.StatusCode != http.StatusOK {
		log.Error("Xentral API returned an error", zap.Int("status_code", res.StatusCode))
		return nil, fmt.Errorf("xentral API returned an error: %d", res.StatusCode)
	}

	if err != nil {
		log.Error("Failed to make GET request", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

func ProcessSalesOrderDetails(order model.Order, log *zap.Logger) error {
	log.Info("Processing sales order details started")

	// Step 1: Get Customer ID from Xentral API
	customerID, err := getCustomerID(order.Email, log)
	if err != nil {
		log.Error("Failed to get customer ID", zap.Error(err))
		return err
	}

	// Step 2: Get Product ID from Xentral API
	productID, err := getProductID(order.Products[0].ProductID, log)
	if err != nil {
		log.Error("Failed to get product ID", zap.Error(err))
		return err
	}

	// Step 3: Set Payment Method ID based on order.PaymentMethod
	var paymentMethodID string
	switch order.PaymentMethod {
	case "card":
		paymentMethodID = "15"
	case "bank":
		paymentMethodID = "13"
	default:
		log.Warn("Unknown payment method, using default ID")
		paymentMethodID = "1" // Use a default ID or handle as needed
	}

	// Step 4: Make Sales Order API Request
	url := "https://652fcc41115ea.xentral.biz/api/salesOrders/actions/import"
	currentDate := time.Now().Format("2006-01-02")

	// Update the payload with the new customerID and productID
	payloadMap := map[string]interface{}{
		"customer": map[string]interface{}{
			"id": customerID,
		},
		"project": map[string]interface{}{
			"id": "1",
		},
		"financials": map[string]interface{}{
			"paymentMethod": map[string]interface{}{
				"id": paymentMethodID,
			},
			"billingAddress": map[string]interface{}{
				"street":  order.InvoiceAddress,
				"country": "CH",
				"name":    order.FirstName + " " + order.LastName,
				"city":    order.InvoiceAddress,
				"zipCode": order.InvoiceAddress,
				"type":    "mr",
			},
			"currency": "CHF",
		},
		"delivery": map[string]interface{}{
			"shippingAddress": map[string]interface{}{
				"street":  order.DeliveryAddress,
				"type":    "mr",
				"name":    order.FirstName + " " + order.LastName,
				"zipCode": order.DeliveryAddress,
				"city":    order.DeliveryAddress,
				"country": "CH",
			},
			"shippingMethod": map[string]interface{}{
				"id": "1",
			},
		},
		"date": currentDate,
		"positions": []map[string]interface{}{
			{
				"product": map[string]interface{}{
					"id": productID,
				},
				"price": map[string]interface{}{
					"amount":   fmt.Sprintf("%.2f", order.TotalPrice), // Format the amount as a string
					"currency": "CHF",
				},
				"quantity": order.Products[0].Quantity,
			},
		},
		"externalOrderId": order.Products[0].OrderID,
	}

	// Convert the payload to JSON
	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		log.Error("Failed to marshal payload to JSON", zap.Error(err))
		return err
	}

	// Log the payload to check the format and values
	log.Info("Sales Order Payload:", zap.String("payload", string(payloadJSON)))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payloadJSON))
	req.Header.Add("accept", "text/html")
	req.Header.Add("content-type", "application/vnd.xentral.default.v1-beta+json")
	req.Header.Add("authorization", "Bearer 9|X63RhGycfCWFIV8FmNZw8YliZ11Vcs9Np99k9VT8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Failed to make sales order API request", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	log.Info("Response from sales order details API:", zap.String("response", string(body)))

	return nil
}
