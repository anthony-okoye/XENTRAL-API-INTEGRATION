package query

import (
	"bookbox-backend/internal/model"
	"bookbox-backend/internal/request"
	"encoding/json"
	"strings"
	"unicode"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CutArrayToLimit(entity string, limit int, rows any) (output any) {
	switch entity {
	case "product":
		val, _ := rows.(*[]model.Product)
		if val == nil {
			return rows
		}
		res := (*val)[0:limit]
		output = &res
	case "review":
		val, _ := rows.(*[]model.Review)
		if val == nil {
			return rows
		}

		res := (*val)[0:limit]
		output = &res

	case "order":
		val, _ := rows.(*[]model.Order)
		if val == nil {
			return rows
		}

		res := (*val)[0:limit]
		output = &res

	case "user":
		val, _ := rows.(*[]model.User)
		if val == nil {
			return rows
		}

		res := (*val)[0:limit]
		output = &res

	case "discount":
		val, _ := rows.(*[]model.Discount)
		if val == nil {
			return rows
		}

		res := (*val)[0:limit]
		output = &res

	case "category":
		val, _ := rows.(*[]model.Category)
		if val == nil {
			return rows
		}

		res := (*val)[0:limit]
		output = &res

	case "sales_channel":
		val, _ := rows.(*[]model.SalesChannel)
		if val == nil {
			return rows
		}

		res := (*val)[0:limit]
		output = &res
	case "cart":
		val, _ := rows.(*[]model.Cart)
		if val == nil {
			return rows
		}

		res := (*val)[0:limit]
		output = &res

	}

	return output
}

func DetermineCount(entity string, rows any) (count int) {
	switch entity {
	case "product":
		val, _ := rows.(*[]model.Product)
		count = len(*val)
	case "review":
		val, _ := rows.(*[]model.Review)
		count = len(*val)

	case "order":
		val, _ := rows.(*[]model.Order)
		count = len(*val)

	case "user":
		val, _ := rows.(*[]model.User)
		count = len(*val)

	case "discount":
		val, _ := rows.(*[]model.Discount)
		count = len(*val)

	case "category":
		val, _ := rows.(*[]model.Category)
		count = len(*val)

	case "sales_channel":
		val, _ := rows.(*[]model.SalesChannel)
		count = len(*val)
	case "cart":
		val, _ := rows.(*[]model.Cart)
		count = len(*val)

	}

	return
}

func Determine(entity string) interface{} {
	switch entity {
	case "product":
		return &model.Product{}

	case "review":
		return &model.Review{}

	case "order":
		return &model.Order{}

	case "user":
		return &model.User{}

	case "discount":
		return &model.Discount{}

	case "category":
		return &model.Category{}

	case "sales_channel":
		return &model.SalesChannel{}
	case "cart":
		return &model.Cart{}

	}

	return nil
}

func DetermineArray(entity string) interface{} {
	switch entity {
	case "product":
		return &[]model.Product{}

	case "review":
		return &[]model.Review{}

	case "order":
		return &[]model.Order{}

	case "user":
		return &[]model.User{}

	case "discount":
		return &[]model.Discount{}

	case "category":
		return &[]model.Category{}

	case "sales_channel":
		return &[]model.SalesChannel{}
	case "cart":
		return &[]model.Cart{}

	}

	return nil
}

func GetID(data any, entity string) (id string) {
	var res any
	switch entity {
	case "product":
		res = *data.(*model.Product)

	case "review":
		res = *data.(*model.Review)

	case "order":
		res = *data.(*model.Order)

	case "user":
		res = *data.(*model.User)

	case "discount":
		res = *data.(*model.Discount)

	case "category":
		res = *data.(*model.Category)

	case "sales_channel":
		res = *data.(*model.SalesChannel)

	case "cart":
		res = *data.(*model.Cart)

	default:
		return ""
	}
	raw, _ := json.Marshal(res)
	dataMap := make(map[string]any)
	json.Unmarshal(raw, &dataMap)

	return dataMap["id"].(string)
}

func GetPreloadMapping(key string) (preloadKey string) {
	splits := strings.Split(key, ".")
	results := make([]string, 0)
	for _, split := range splits {
		split = strings.ToLower(split)

		var result strings.Builder
		capitalizeNext := true
		for _, c := range split {
			if c == '_' {
				capitalizeNext = true
			} else if capitalizeNext {
				result.WriteRune(unicode.ToUpper(c))
				capitalizeNext = false
			} else {
				result.WriteRune(c)
			}
		}
		results = append(results, result.String())
	}

	return strings.Join(results, ".")
}

func DetermineRelations(listRequest request.GetRequest, db *gorm.DB) *gorm.DB {
	for _, relationship := range listRequest.Metadata.Relationships {
		if relationship.Name == "*" {
			db = db.Preload(clause.Associations)
			break
		}

		if len(relationship.RelationParams) == 0 {
			key := GetPreloadMapping(relationship.Name)
			db = db.Preload(key)
			continue
		}

		db = Relate(listRequest.Entity, relationship, db)
	}

	return db
}

func Relate(entity string, relation request.Relationship, db *gorm.DB) *gorm.DB {
	key := GetPreloadMapping(relation.Name)
	db = db.Preload(key)

	switch entity {
	case "product":
		switch relation.Name {
		case "sales_channels":
			where := makeCondition(relation.RelationParams, "sales_channels", false)
			db = db.
				Joins("JOIN sales_channel_products ON products.id = sales_channel_products.product_id").
				Joins("JOIN sales_channels ON sales_channels.id = sales_channel_products.sales_channel_id").
				Where(where.Main, where.Values...)

		case "categories":
			where := makeCondition(relation.RelationParams, "categories", false)
			db = db.
				Joins("JOIN product_categories ON products.id = product_categories.product_id").
				Joins("JOIN categories ON categories.id = product_categories.category_id").
				Where(where.Main, where.Values...)
		}
	case "review":
		switch relation.Name {
		case "user":
			where := makeCondition(relation.RelationParams, "users", false)
			db = db.
				Joins("JOIN users ON reviews.user_id = users.id").
				Where(where.Main, where.Values...)
		case "product":
			where := makeCondition(relation.RelationParams, "products", false)
			db = db.
				Joins("JOIN products ON reviews.product_id = product.id").
				Where(where.Main, where.Values...)
		}
	case "order":
		switch relation.Name {
		case "products":
			where := makeCondition(relation.RelationParams, "products", false)
			db = db.
				Joins("JOIN order_products ON orders.id = order_products.order_id").
				Joins("JOIN products ON products.id = order_products.product_id").
				Where(where.Main, where.Values...)

		case "user":
			where := makeCondition(relation.RelationParams, "users", false)
			db = db.
				Joins("JOIN users ON orders.user_id = users.id").
				Where(where.Main, where.Values...)

		case "sales_channel":
			where := makeCondition(relation.RelationParams, "sales_channels", false)
			db = db.
				Joins("JOIN sales_channels ON orders.sales_channel_id = sales_channels.id").
				Where(where.Main, where.Values...)
		}
	case "user":
	case "discount":
		switch relation.Name {
		case "sales_channel":
			where := makeCondition(relation.RelationParams, "sales_channels", false)
			db = db.
				Joins("JOIN discount_sales_channels ON discounts.id = discount_sales_channels.discount_id").
				Joins("JOIN sales_channels ON sales_channels.id = discount_sales_channels.sales_channel_id").
				Where(where.Main, where.Values...)
		case "user":
			where := makeCondition(relation.RelationParams, "users", false)
			db = db.
				Joins("JOIN users ON discounts.user_id = users.id").
				Where(where.Main, where.Values...)

		}
	case "category":
		switch relation.Name {
		case "sub_categories":
			where := makeCondition(relation.RelationParams, "c", false)
			db = db.
				Joins("JOIN category_sub_categories ON categories.id = category_sub_categories.category_id").
				Joins("JOIN categories c ON c.id = category_sub_categories.sub_category_id").
				Where(where.Main, where.Values...)
		}
	case "sales_channel":
		switch relation.Name {
		case "discounts":
			where := makeCondition(relation.RelationParams, "discounts", false)
			db = db.
				Joins("JOIN discount_sales_channels ON sales_channels.id = discount_sales_channels.sales_channel_id").
				Joins("JOIN discounts ON discounts.id = discount_sales_channels.discount_id").
				Where(where.Main, where.Values...)

		}
	}

	return db
}
