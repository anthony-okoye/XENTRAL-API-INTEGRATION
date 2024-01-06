package model

import (
	"encoding/base64"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	Root
	Title           string                `json:"title,omitempty" gorm:"column:title;"`
	Subtitle        string                `json:"subtitle,omitempty" gorm:"column:subtitle"`
	Description     string                `json:"description,omitempty" gorm:"column:description"`
	ISBN            string                `json:"isbn,omitempty" gorm:"column:isbn"`
	EAN             string                `json:"ean,omitempty" gorm:"column:ean"`
	BZNR            string                `json:"bznr,omitempty" gorm:"column:bznr"`
	CoverPicture    string                `json:"cover_picture,omitempty" gorm:"column:cover_picture"`
	PublicationDate int64                 `json:"publication_date,omitempty" gorm:"column:publication_date"`
	Edition         string                `json:"edition,omitempty" gorm:"column:edition"`
	Publisher       string                `json:"publisher,omitempty" gorm:"column:publisher"`
	Stock           int                   `json:"stock" gorm:"column:stock"`
	DeliveryTime    string                `json:"delivery_time,omitempty" gorm:"column:delivery_time"`
	SellingPrice    float64               `json:"selling_price" gorm:"column:selling_price"`
	Language        string                `json:"language" gorm:"column:language"`
	Width           string                `json:"width,omitempty" gorm:"column:width"`
	Height          string                `json:"height,omitempty" gorm:"column:height"`
	Length          string                `json:"length,omitempty" gorm:"column:length"`
	Weight          string                `json:"weight,omitempty" gorm:"column:weight"`
	Replacement     string                `json:"replacement,omitempty" gorm:"column:replacement"`
	IsDownloadTitle bool                  `json:"is_download_title" gorm:"column:is_download_title"`
	Reviews         []Review              `json:"reviews,omitempty" gorm:"foreignKey:product_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Categories      []ProductCategory     `json:"categories,omitempty" gorm:"foreignKey:product_id;constraint:OnDelete:SET NULL;"`
	SalesChannels   []SalesChannelProduct `json:"sales_channels,omitempty" gorm:"foreignKey:product_id;constraint:OnDelete:SET NULL;"`
}

type ProductCategory struct {
	ID         string   `json:"id" gorm:"primaryKey"`
	ProductID  string   `json:"product_id"`
	CategoryID string   `json:"category_id"`
	Category   Category `json:"category" gorm:"foreignKey:category_id;constraint:OnDelete:SET NULL;"`
	Order      int      `json:"order"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if len(p.ID) == 0 {
		id := uuid.New().String()
		p.ID = id
	}

	if p.Active == nil {
		value := true
		p.Active = &value
	}

	if p.CoverPicture != "" && strings.Contains(p.CoverPicture, "base64,") {
		var err error

		// split needed because it comes in this format: data:image/png;base64,iVBORw0KGgoAAAA...
		splits := strings.Split(p.CoverPicture, "base64,")
		p.CoverPicture = splits[1]

		p.CoverPicture, err = UploadImage(imageProductRootPath, p.ID, p.CoverPicture)
		if err != nil {
			return err
		}
	}
	p.CreatedAt = time.Now()

	return nil
}

const imageProductRootPath = "/tmp/bookbox-api/images/product"

func (p *Product) BeforeUpdate(tx *gorm.DB) error {
	if p.CoverPicture != "" && strings.Contains(p.CoverPicture, "base64,") {
		var err error

		// split needed because it comes in this format: data:image/png;base64,iVBORw0KGgoAAAA...
		splits := strings.Split(p.CoverPicture, "base64,")
		p.CoverPicture = splits[1]

		p.CoverPicture, err = UploadImage(imageProductRootPath, p.ID, p.CoverPicture)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Product) AfterUpdate(tx *gorm.DB) error {
	if tx.Error != nil {
		os.Remove(p.CoverPicture)
	}

	return nil
}

func (p *Product) AfterCreate(tx *gorm.DB) error {
	if tx.Error != nil {
		os.Remove(p.CoverPicture)
	}

	return nil
}

func (p *Product) AfterFind(tx *gorm.DB) error {
	raw, _ := os.ReadFile(p.CoverPicture)
	if len(raw) != 0 {
		p.CoverPicture = base64.StdEncoding.EncodeToString([]byte(raw))
	}
	return nil
}

func (p *Product) AfterDelete(tx *gorm.DB) error {
	os.Remove(p.CoverPicture)

	return nil
}

func (pc *ProductCategory) BeforeCreate(tx *gorm.DB) error {
	if len(pc.ID) == 0 {
		id := uuid.New().String()
		pc.ID = id
	}

	return nil
}
