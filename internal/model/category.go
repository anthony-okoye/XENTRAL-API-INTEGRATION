package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	Root
	Name          string            `json:"name,omitempty" gorm:"column:name"`
	URL           string            `json:"url,omitempty" gorm:"column:url"`
	IsRoot        bool              `json:"is_root,omitempty" gorm:"column:is_root"`
	ParentID      *string           `json:"parent_id" gorm:"column:parent_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Parent        *Category         `json:"parent" gorm:"foreignkey:parent_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SubCategories []*Category       `json:"sub_categories,omitempty" gorm:"many2many:category_sub_categories;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Products      []ProductCategory `json:"products,omitempty" gorm:"foreignKey:category_id;constraint:OnDelete:SET NULL;"`
}

type CategorySubCategory struct {
	CategoryID    string `gorm:"primaryKey;column:category_id"`
	SubCategoryID string `gorm:"primaryKey;column:sub_category_id"`
}

func (c *CategorySubCategory) BeforeSave(tx *gorm.DB) error {

	category := Category{}
	category.ID = c.SubCategoryID
	category.ParentID = &c.CategoryID

	err := tx.Model(&category).Update("parent_id", c.CategoryID).Error
	if err != nil {
		return err
	}

	return nil
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
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
