package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	UserCustomerRole = "customer"
	UserAdminRole    = "admin"
)

type User struct {
	Root
	Role              string     `json:"role,omitempty" gorm:"type:user_role;not null;column:role"`
	Password          string     `json:"password" gorm:"column:password"`
	Salutation        string     `json:"salutation,omitempty" gorm:"type:user_salutation;not null;column:salutation"`
	FirstName         string     `json:"first_name,omitempty" gorm:"column:first_name"`
	Type              string     `json:"type" gorm:"type:user_type;not null;column:type"`
	LastName          string     `json:"last_name,omitempty" gorm:"column:last_name"`
	ZipCode           string     `json:"zip_code,omitempty" gorm:"column:zip_code"`
	City              string     `json:"city,omitempty" gorm:"column:city"`
	Email             string     `json:"email,omitempty" gorm:"column:email"`
	PhoneNumber       string     `json:"phone_number,omitempty" gorm:"column:phone_number"`
	Country           string     `json:"country,omitempty" gorm:"column:country"`
	Reviews           []Review   `json:"reviews,omitempty" gorm:"foreignKey:user_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Discounts         []Discount `json:"discounts,omitempty" gorm:"foreignKey:user_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Orders            []Order    `json:"orders,omitempty" gorm:"foreignKey:user_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Favorites         []Favorite `json:"favorites,omitempty" gorm:"foreignKey:user_id;constraint:OnDelete:CASCADE"`
	DeliveryAddressID *string    `json:"delivery_address_id,omitempty"`
	DeliveryAddress   *Address   `json:"delivery_address,omitempty" gorm:"foreignKey:delivery_address_id;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	BillingAddressID  *string    `json:"billing_address_id,omitempty"`
	BillingAddress    *Address   `json:"billing_address,omitempty" gorm:"foreignKey:billing_address_id;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Favorite struct {
	ID             string       `json:"id" gorm:"primaryKey"`
	UserID         string       `json:"user_id"`
	ProductID      string       `json:"product_id"`
	Product        Product      `json:"product" gorm:"foreignKey:product_id;constraint:OnDelete:CASCADE"`
	SalesChannelID string       `json:"sales_channel_id"`
	SalesChannel   SalesChannel `json:"sales_channel" gorm:"foreignKey:sales_channel_id;constraint:OnDelete:CASCADE"`
}

func (f *Favorite) BeforeCreate(tx *gorm.DB) error {
	if len(f.ID) == 0 {
		id := uuid.New().String()
		f.ID = id
	}

	return nil
}

func (c *Favorite) BeforeUpdate(tx *gorm.DB) error {
	tx.Model(c).Association("Favorites").Clear()

	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if len(u.ID) == 0 {
		id := uuid.New().String()
		u.ID = id
	}

	if u.Active == nil {
		value := true
		u.Active = &value
	}

	u.CreatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if u.BillingAddress == nil {
		u.BillingAddressID = nil
	}

	if u.DeliveryAddress == nil {
		u.DeliveryAddressID = nil
	}

	var user User

	result := tx.Preload(clause.Associations).Where("id = ?", u.ID).First(&user)
	if result.Error != nil {
		return result.Error
	}

	if user.BillingAddress != nil {
		if err := tx.Delete(&Address{}, "id = ?", user.BillingAddress.ID).Error; err != nil {
			return err
		}
	}

	if user.DeliveryAddress != nil {
		if err := tx.Delete(&Address{}, "id = ?", user.DeliveryAddress.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

type Address struct {
	Root
	Address string `json:"address,omitempty" gorm:"column:address"`
	Country string `json:"country,omitempty" gorm:"column:country"`
	Street  string `json:"street,omitempty" gorm:"column:street"`
	ZipCode string `json:"zip_code,omitempty" gorm:"column:zip_code"`
	City    string `json:"city,omitempty" gorm:"column:city"`
}

func (a *Address) BeforeCreate(tx *gorm.DB) error {
	if len(a.ID) >= 1 {
		return nil
	}

	id := uuid.New().String()
	a.ID = id

	return nil
}
