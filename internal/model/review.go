package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Review struct {
	Root
	Title     string  `json:"title,omitempty" gorm:"column:title"`
	Stars     int     `json:"stars,omitempty" gorm:"column:stars"`
	Date      int64   `json:"date,omitempty" gorm:"column:date"`
	Comment   string  `json:"comment,omitempty" gorm:"column:comment"`
	UserID    string  `json:"user_id,omitempty" gorm:"column:user_id"`
	User      User    `json:"user,omitempty" gorm:"foreignKey:user_id"`
	ProductID string  `json:"product_id,omitempty" gorm:"column:product_id"`
	Product   Product `json:"product,omitempty" gorm:"foreignKey:product_id"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if len(r.ID) == 0 {
		id := uuid.New().String()
		r.ID = id
	}

	if r.Active == nil {
		value := true
		r.Active = &value
	}

	r.Date = time.Now().Unix()
	r.CreatedAt = time.Now()

	return nil
}
