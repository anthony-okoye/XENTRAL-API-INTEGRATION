package model

import (
	"encoding/base64"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesChannel struct {
	Root
	Name         string                `json:"name,omitempty" gorm:"column:name"`
	Domain       string                `json:"domain,omitempty" gorm:"column:domain"`
	DescTitle    string                `json:"desc_title,omitempty" gorm:"column:desc_title"`
	Description  string                `json:"description,omitempty" gorm:"column:description"`
	CoverPicture string                `json:"cover_picture,omitempty" gorm:"column:cover_picture"`
	Products     []SalesChannelProduct `json:"products,omitempty" gorm:"foreignKey:sales_channel_id;constraint:OnDelete:CASCADE"`
	Discounts    []*Discount           `json:"discounts,omitempty" gorm:"many2many:discount_sales_channels; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Categories   []Category            `json:"categories,omitempty" gorm:"many2many:sales_channel_categories;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Orders       []Order               `json:"orders,omitempty" gorm:"foreignKey:sales_channel_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type SalesChannelProduct struct {
	ID             string  `json:"id" gorm:"primaryKey"`
	SalesChannelID string  `json:"sales_channel_id"`
	ProductID      string  `json:"product_id"`
	Product        Product `json:"product" gorm:"foreignKey:product_id;constraint:OnDelete:SET NULL;"`
	ChangedPrice   float64 `json:"changed_price"`
	ChangedTitle   string  `json:"changed_title"`
}

func (pc *SalesChannelProduct) BeforeCreate(tx *gorm.DB) error {
	if len(pc.ID) == 0 {
		id := uuid.New().String()
		pc.ID = id
	}

	return nil
}

func (sc *SalesChannel) BeforeCreate(tx *gorm.DB) error {
	if len(sc.ID) == 0 {
		id := uuid.New().String()
		sc.ID = id
	}

	if sc.Active == nil {
		value := true
		sc.Active = &value
	}

	if sc.CoverPicture != "" && strings.Contains(sc.CoverPicture, "base64,") {
		var err error

		// split needed because it comes in this format: data:image/png;base64,iVBORw0KGgoAAAA...
		splits := strings.Split(sc.CoverPicture, "base64,")
		sc.CoverPicture = splits[1]

		sc.CoverPicture, err = UploadImage(imageSCRootPath, sc.ID, sc.CoverPicture)
		if err != nil {
			return err
		}
	}
	sc.CreatedAt = time.Now()

	return nil
}

const imageSCRootPath = "/tmp/bookbox-api/images/sales_channel"

func (sc *SalesChannel) BeforeUpdate(tx *gorm.DB) error {
	if sc.CoverPicture != "" && strings.Contains(sc.CoverPicture, "base64,") {
		var err error

		// split needed because it comes in this format: data:image/png;base64,iVBORw0KGgoAAAA...
		splits := strings.Split(sc.CoverPicture, "base64,")
		sc.CoverPicture = splits[1]

		sc.CoverPicture, err = UploadImage(imageSCRootPath, sc.ID, sc.CoverPicture)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *SalesChannel) AfterUpdate(tx *gorm.DB) error {
	if tx.Error != nil {
		os.Remove(sc.CoverPicture)
	}

	return nil
}

func (sc *SalesChannel) AfterCreate(tx *gorm.DB) error {
	if tx.Error != nil {
		os.Remove(sc.CoverPicture)
	}

	return nil
}

func (sc *SalesChannel) AfterFind(tx *gorm.DB) error {
	raw, _ := os.ReadFile(sc.CoverPicture)
	if len(raw) != 0 {
		sc.CoverPicture = base64.StdEncoding.EncodeToString([]byte(raw))
	}
	return nil
}

func (sc *SalesChannel) AfterDelete(tx *gorm.DB) error {
	os.Remove(sc.CoverPicture)

	return nil
}
