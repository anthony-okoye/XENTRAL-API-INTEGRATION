package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Discount struct {
	Root
	Name          string          `json:"name,omitempty" gorm:"column:name"`
	ValidFrom     string          `json:"valid_from,omitempty" gorm:"column:valid_from"`
	ValidTo       string          `json:"valid_to,omitempty" gorm:"column:valid_to"`
	Count         int             `json:"count,omitempty" gorm:"column:count"`
	CountPerUser  int             `json:"count_per_user,omitempty" gorm:"column:count_per_user"`
	Percent       float64         `json:"percent,omitempty" gorm:"column:percent"`
	SalesChannels []*SalesChannel `json:"sales_channels,omitempty" gorm:"many2many:discount_sales_channels;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID        string          `json:"user_id,omitempty" gorm:"column:user_id"`
	User          User            `json:"user,omitempty" gorm:"foreignKey:user_id"`
}

func (d *Discount) BeforeCreate(tx *gorm.DB) error {
	if len(d.ID) == 0 {
		id := uuid.New().String()
		d.ID = id
	}

	if d.Active == nil {
		value := true
		d.Active = &value
	}
	d.CreatedAt = time.Now()

	return nil
}
