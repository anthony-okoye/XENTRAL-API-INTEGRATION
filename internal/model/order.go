package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	Root
	FirstName       string        `json:"first_name,omitempty" gorm:"column:first_name"`
	LastName        string        `json:"last_name,omitempty" gorm:"column:last_name"`
	Number          int           `json:"number,omitempty" gorm:"column:number"`
	Email           string        `json:"email,omitempty" gorm:"column:email"`
	InvoiceAddress  string        `json:"invoice_address,omitempty" gorm:"column:invoice_address"`
	TotalPrice      float64       `json:"total_price" gorm:"column:total_price"`
	Status          string        `json:"status,omitempty" gorm:"column:status"`
	PaymentStatus   string        `json:"payment_status,omitempty" gorm:"type:payment_status;not null;column:payment_status"`
	PaymentMethod   string        `json:"payment_method,omitempty" gorm:"type:payment_method;not null;column:payment_method"`
	ShipmentNumber  int           `json:"shipment_number,omitempty" gorm:"column:shipment_number"`
	OrderStatus     string        `json:"order_status,omitempty" gorm:"type:order_status;not null;column:order_status"`
	DeliveryStatus  string        `json:"delivery_status,omitempty" gorm:"type:delivery_status;not null;column:delivery_status"`
	DeliveryAddress string        `json:"delivery_address,omitempty" gorm:"column:delivery_address"`
	Date            int64         `json:"date,omitempty" gorm:"column:date"`
	EBookOrderID    string        `json:"eBook_order_id" gorm:"column:eBook_order_id"`
	SalesChannelID  *string       `json:"sales_channel_id,omitempty" gorm:"column:sales_channel_id"`
	SalesChannel    *SalesChannel `json:"sales_channel,omitempty" gorm:"foreignKey:sales_channel_id"`
	Products        []OrderItem   `json:"products,omitempty" gorm:"foreignKey:order_id;constraint:OnDelete:CASCADE"`
	UserID          *string       `json:"user_id,omitempty" gorm:"column:user_id"`
	User            *User         `json:"user,omitempty" gorm:"foreignKey:user_id"`
}

type OrderItem struct {
	ID           string  `json:"id" gorm:"primaryKey"`
	OrderID      string  `json:"order_id"`
	ProductID    string  `json:"product_id"`
	Product      Product `json:"product" gorm:"foreignKey:product_id;constraint:OnDelete:CASCADE"`
	EBookURL     string  `json:"ebook_url" gorm:"column:eBook_url"`
	Quantity     int     `json:"quantity"`
	CurrentPrice float64 `json:"current_price"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if len(o.ID) == 0 {
		id := uuid.New().String()
		o.ID = id
	}

	if o.Active == nil {
		value := true
		o.Active = &value
	}

	o.Date = time.Now().Unix()
	o.CreatedAt = time.Now()
	return nil
}

func (o *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if len(o.ID) == 0 {
		id := uuid.New().String()
		o.ID = id
	}

	return nil
}

type OrderNotification struct {
	Details     Order `json:"details"`
	EbooksCount int   `json:"ebooksCount"`
	RetryCount  int   `json:"retryCount"`
}
