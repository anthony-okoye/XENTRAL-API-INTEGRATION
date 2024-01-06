package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Cart struct {
	Root
	CartItems []CartItem `json:"cart_items,omitempty" gorm:"foreignKey:cart_id;constraint:OnDelete:CASCADE"`
}

type CartItem struct {
	ID        string  `json:"id" gorm:"primaryKey"`
	CartID    string  `json:"cart_id"`
	ProductID string  `json:"product_id"`
	Product   Product `json:"product" gorm:"foreignKey:product_id;constraint:OnDelete:CASCADE"`
	Quantity  int     `json:"quantity"`
}

func (c *Cart) BeforeCreate(tx *gorm.DB) error {
	if len(c.ID) == 0 {
		id := uuid.New().String()
		c.ID = id
	}

	if c.Active == nil {
		value := true
		c.Active = &value
	}
	c.CreatedAt = time.Now()
	return nil
}

func (c *CartItem) BeforeCreate(tx *gorm.DB) error {
	if len(c.ID) == 0 {
		id := uuid.New().String()
		c.ID = id
	}

	return nil
}

func (c *CartItem) BeforeUpdate(tx *gorm.DB) error {
	tx.Model(c).Association("CartItems").Clear() // clear all cart items

	return nil
}
