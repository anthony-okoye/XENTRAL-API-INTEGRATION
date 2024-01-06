package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() (err error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", database.Host, database.User, database.Password, database.Database, database.Port)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}

	// Set the maximum number of open connections
	sqlDB.SetMaxOpenConns(100) // Adjust based on your requirements

	// Set the maximum number of idle connections
	sqlDB.SetMaxIdleConns(10) // Adjust based on your requirements

	// Set the maximum connection lifetime
	sqlDB.SetConnMaxLifetime(time.Hour) // Adjust based on your requirements

	// Run migrations.
	Migrate(gormDB)
	if err != nil {
		fmt.Println("failed to migrate :", err.Error())
	}

	DB = gormDB

	// pg trgrm
	err = DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm;").Error
	if err != nil {
		fmt.Println("failed to create extension pg_trgm :", err.Error())
	}

	// Execute the raw SQL query to create an index
	if err = DB.Exec("CREATE INDEX products_isbn_search ON products (isbn)").Error; err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// search index for title
	err = DB.Exec("CREATE INDEX products_title_trgm_idx ON products USING gin(title gin_trgm_ops)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// search index for publisher
	err = DB.Exec("CREATE INDEX products_publisher_trgm_idx ON products USING gin(publisher gin_trgm_ops)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// index for product_categories
	err = DB.Exec("CREATE INDEX idx_product_categories_product_id ON product_categories (product_id)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// index for product_categories
	err = DB.Exec("CREATE INDEX idx_product_categories_category_id ON product_categories (category_id)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// index for sales_channel_products
	err = DB.Exec("CREATE INDEX idx_sales_channel_product_sales_channel_id ON sales_channel_products (sales_channel_id)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// index for sales_channel_products
	err = DB.Exec("CREATE INDEX idx_sales_channel_product_product_id ON sales_channel_products (product_id)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// index for categories name
	err = DB.Exec("CREATE INDEX idx_categories_name ON categories (name)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// sort index for title
	err = DB.Exec("CREATE INDEX products_title_index ON products(title)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// sort index for publisher
	err = DB.Exec("CREATE INDEX products_publisher_index ON products(publisher)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// sort index for pub date
	err = DB.Exec("CREATE INDEX products_pubdate_index ON products(publication_date)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// sort index for selling price
	err = DB.Exec("CREATE INDEX products_selling_price_index ON products(selling_price)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// sort index for selling price
	err = DB.Exec("CREATE INDEX products_stock_index ON products(stock)").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET effective_io_concurrency = '200';").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET work_mem = '640MB';").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET shared_buffers = '8GB';").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET effective_cache_size = '16GB';").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET maintenance_work_mem = '2GB';").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET max_connections = '200';").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET enable_seqscan = on;").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	err = DB.Exec("ALTER SYSTEM SET plan_cache_mode = force_custom_plan;").Error
	if err != nil {
		fmt.Println("failed to create index :", err.Error())
	}

	// Set the Autovacuum parameters
	DB.Exec("ALTER SYSTEM SET autovacuum_max_workers TO 6;")
	DB.Exec("ALTER SYSTEM SET maintenance_work_mem TO '2000MB';")
	DB.Exec("ALTER SYSTEM SET autovacuum_vacuum_scale_factor TO 0.01;")

	err = DB.Exec("SELECT pg_reload_conf();").Error
	if err != nil {
		fmt.Println("failed to reload configuration:", err.Error())
	}

	return nil
}

func omitPassword(db *gorm.DB) {

	db.Omit("password")
}
