package database

import (
	"bookbox-backend/internal/model"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed enums.sql
var enums []byte

func Migrate(gormDB *gorm.DB) (err error) {
	err = gormDB.Exec(string(enums)).Error
	if err != nil {
		return
	}

	err = gormDB.SetupJoinTable(&model.Category{}, "SubCategories", &model.CategorySubCategory{})
	if err != nil {
		return err
	}

	// Migrate ORM models.
	err = gormDB.AutoMigrate(
		&model.Category{},
		&model.CategorySubCategory{},
		&model.Product{},
		&model.ProductCategory{},
		&model.Cart{},
		&model.CartItem{},
		&model.SalesChannel{},
		&model.SalesChannelProduct{},
		&model.User{},
		&model.Favorite{},
		&model.Review{},
		&model.Order{},
		&model.OrderItem{},
		&model.Discount{},
		&model.Address{},
		&model.Sync{},
	)
	if err != nil {
		return
	}

	//Seed()

	return
}
