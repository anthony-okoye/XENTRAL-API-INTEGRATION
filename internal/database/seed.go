package database

import (
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/crypto"
	"bookbox-backend/pkg/logger"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	stressTest = 10
)

func Seed() (err error) {

	log := logger.Log
	log.Info("seeding started")

	err = SeedSyncData(log)
	if err != nil {
		log.Error("failed to seed Sync data", zap.Error(err))
	}

	err = SeedUsers(log)
	if err != nil {
		log.Error("failed to seed Users", zap.Error(err))
	}

	err = SeedSalesChannels(log)
	if err != nil {
		log.Error("failed to seed sales channels", zap.Error(err))
	}

	// err = SeedOrders(log)
	// if err != nil {
	// 	return
	// }

	log.Info("seeding finished")

	return
}

func SeedSyncData(log *zap.Logger) (err error) {
	syncData := model.Sync{
		Root: model.Root{
			ID: "1",
		},
		IsFullSynced: false,
	}

	DB.Create(&syncData)

	log.Info("seeding finished")

	return
}

func SeedUsers(log *zap.Logger) (err error) {
	user := model.User{
		FirstName:  "Name1",
		LastName:   "Surname1",
		Email:      "user1@bookbox.com",
		Password:   "Password1",
		Role:       model.UserCustomerRole,
		Salutation: "Frau",
		Type:       "Privat",
		DeliveryAddress: &model.Address{
			Address: "delivery",
		},
		BillingAddress: &model.Address{
			Address: "billing",
		},
	}
	user.ID = "1"
	user.Password = crypto.HashPassword(user.Password)
	err = DB.Create(&user).Error
	if err != nil {
		return
	}

	user = model.User{
		FirstName:  "Name2",
		LastName:   "Surname2",
		Email:      "user2@bookbox.com",
		Password:   "Password2",
		Role:       model.UserCustomerRole,
		Salutation: "Frau",
		Type:       "Privat",
		DeliveryAddress: &model.Address{
			Address: "delivery",
		},
		BillingAddress: &model.Address{
			Address: "billing",
		},
	}
	user.ID = "2"

	user.Password = crypto.HashPassword(user.Password)
	err = DB.Create(&user).Error
	if err != nil {
		return
	}

	user = model.User{
		FirstName:  "Name3",
		LastName:   "Surname3",
		Email:      "admin@bookbox.com",
		Password:   "admin",
		Role:       model.UserAdminRole,
		Salutation: "Frau",
		Type:       "Privat",
		DeliveryAddress: &model.Address{
			Address: "delivery",
		},
		BillingAddress: &model.Address{
			Address: "billing",
		},
	}
	user.ID = "3"

	user.Password = crypto.HashPassword(user.Password)
	err = DB.Create(&user).Error
	if err != nil {
		return
	}

	return
}

func SeedDiscounts(log *zap.Logger) (err error) {

	discount := model.Discount{
		Name:      "Discount1",
		Root:      model.Root{},
		ValidFrom: time.Now().String(),
		ValidTo:   time.Now().Add(30 * 24 * time.Hour).String(),
		UserID:    strconv.Itoa(1),
	}
	discount.ID = strconv.Itoa(1)

	err = DB.Create(&discount).Error
	if err != nil {
		return
	}

	discount = model.Discount{
		Name:      "Discount2",
		Root:      model.Root{},
		ValidFrom: time.Now().String(),
		ValidTo:   time.Now().Add(30 * 24 * time.Hour).String(),
		UserID:    strconv.Itoa(2),
	}
	discount.ID = strconv.Itoa(2)

	err = DB.Create(&discount).Error
	if err != nil {
		return
	}

	return
}

func SeedSalesChannels(log *zap.Logger) (err error) {
	salesChannel := model.SalesChannel{
		Name:   "BVS Buchverlag API",
		Domain: "bvs",
	}
	salesChannel.ID = "1"
	err = DB.Create(&salesChannel).Error
	if err != nil {
		return
	}

	return
}

func SeedOrders(log *zap.Logger) (err error) {

	productID_1, _ := crypto.UUID()
	productID_2, _ := crypto.UUID()
	userID_1 := "1"
	userID_2 := "2"
	salesChannelID := "1"
	orderID_1, _ := crypto.UUID()
	orderID_2, _ := crypto.UUID()
	orderItemID_1, _ := crypto.UUID()
	orderItemID_2, _ := crypto.UUID()

	product := model.Product{
		Stock:        50,
		SellingPrice: 2,
	}
	product.ID = productID_1

	err = DB.Create(&product).Error
	if err != nil {
		return
	}

	orderItem := model.OrderItem{
		ID:        orderItemID_1,
		OrderID:   orderID_1,
		ProductID: product.ID,
		Quantity:  2,
	}

	order := model.Order{
		UserID: &userID_1,
		Products: []model.OrderItem{
			orderItem,
		},
		Date:           time.Now().Unix(),
		SalesChannelID: &salesChannelID,
		PaymentMethod:  "card",
		PaymentStatus:  "pending",
		DeliveryStatus: "open",
		OrderStatus:    "in_progress",
	}
	order.ID = orderID_1

	err = DB.Create(&order).Error
	if err != nil {
		return
	}

	// err = DB.Create(&orderItem).Error
	// if err != nil {
	// 	return
	// }

	time.Sleep(100 * time.Millisecond)

	product2 := model.Product{
		Stock:        50,
		SellingPrice: 2,
	}
	product2.ID = productID_2

	err = DB.Create(&product2).Error
	if err != nil {
		return
	}

	orderItem2 := model.OrderItem{
		ID:        orderItemID_2,
		OrderID:   orderID_2,
		ProductID: product2.ID,
		Quantity:  2,
	}

	order2 := model.Order{
		UserID: &userID_2,
		Products: []model.OrderItem{
			orderItem2,
		},
		Date:           time.Now().Unix(),
		SalesChannelID: &salesChannelID,
		PaymentMethod:  "card",
		PaymentStatus:  "pending",
		DeliveryStatus: "open",
		OrderStatus:    "in_progress",
	}
	order2.ID = orderID_2

	err = DB.Create(&order2).Error
	if err != nil {
		return
	}

	// err = DB.Create(&orderItem2).Error
	// if err != nil {
	// 	return
	// }

	time.Sleep(100 * time.Millisecond)

	// review := model.Review{
	// 	Title:     "Testing review",
	// 	Comment:   "This book was the best i love it very much!",
	// 	Root:      model.Root{},
	// 	UserID:    userID_1,
	// 	ProductID: productID_1,
	// 	Stars:     5,
	// 	Date:      time.Now().Unix(),
	// }
	// review.ID, _ = crypto.UUID()

	// err = DB.Create(&review).Error
	// if err != nil {
	// 	return
	// }

	// review = model.Review{
	// 	Title:   "Testing review 2",
	// 	Comment: "This book was the best i love it very much! This book was the best i love it very much. This book was the best i love it very much.",

	// 	Root:      model.Root{},
	// 	UserID:    userID_2,
	// 	ProductID: productID_2,
	// 	Stars:     4,
	// 	Date:      time.Now().Unix(),
	// }
	// review.ID, _ = crypto.UUID()

	// err = DB.Create(&review).Error
	// if err != nil {
	// 	return
	// }

	return
}

/*
	Discounts := []*model.Discount{}

	for i := 1; i < 1000; i++ {
		discount := &model.Discount{}
		discount.ID = uint(i)
		discount.UserID = 1

		Discounts = append(Discounts, discount)
	}

	for i := 1; i < 1000; i++ {
		salesChannel := model.SalesChannel{
			Name:      "SalesChannel" + strconv.Itoa(i),
			Products:  Products,
			Discounts: Discounts,
		}
		salesChannel.ID = uint(i)

		err = DB.Create(&salesChannel).Error
		if err != nil {
			return
		}
*/
