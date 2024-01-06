package query

import (
	"bookbox-backend/internal/request"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

var EntityToTableName = map[string]string{
	"product":       "products",
	"category":      "categories",
	"sales_channel": "sales_Channels",
	"review":        "reviews",
	"order":         "orders",
	"user":          "users",
	"discount":      "discounts",
	"cart":          "carts",
}

func countDigits(str string) int {
	count := 0
	for _, ch := range str {
		if unicode.IsDigit(ch) {
			count++
		}
	}
	return count
}

func Constraints(req request.GetRequest) (where request.ConditionStatement, or request.ConditionStatement, err error) {
	//err is unused for now, add if needed
	if len(req.Metadata.Filter.Must) > 0 && strings.ToLower(req.Metadata.Filter.Must[0].Key) == "special" {
		val, ok := req.Metadata.Filter.Must[0].Value.(string)
		if !ok {
			err = fmt.Errorf("when using special in filter the value must be string")
		}

		count := countDigits(val)
		if count >= 10 {
			req.Metadata.Filter.Should = make([]request.FilterParam, 0)
			req.Metadata.Filter.Should = append(req.Metadata.Filter.Should, request.FilterParam{
				Key:   "isbn",
				Value: val,
				Type:  "eq",
			})
		} else {
			req.Metadata.Filter.Should = make([]request.FilterParam, 0)
			req.Metadata.Filter.Should = append(req.Metadata.Filter.Should, request.FilterParam{
				Key:   "title",
				Value: "%" + val + "%",
				Type:  "like",
			})

			req.Metadata.Filter.Should = append(req.Metadata.Filter.Should, request.FilterParam{
				Key:   "publisher",
				Value: "%" + val + "%",
				Type:  "like",
			})
		}

		where.Values = make([]interface{}, 0)
		where.Main = ""

		or = makeCondition(req.Metadata.Filter.Should, EntityToTableName[req.Entity], true)
	} else {
		or.Values = make([]interface{}, 0)
		or.Main = ""

		where = makeCondition(req.Metadata.Filter.Must, EntityToTableName[req.Entity], false)
	}

	return
}

func makeCondition(filters []request.FilterParam, tableName string, isOr bool) (condition request.ConditionStatement) {
	condition.Values = make([]interface{}, 0)
	condition.Main = ""

	for i, filter := range filters {
		filterKey := fmt.Sprintf("%s.%s", tableName, filter.Key)
		filter.Type = strings.ToLower(filter.Type)

		condition.Main += filterKey + " "
		switch filter.Type {
		case "eq":
			condition.Main += "= "
		case "neq":
			condition.Main += "<> "
		case "lt":
			condition.Main += "< "
		case "gt":
			condition.Main += "> "
		case "lte":
			condition.Main += "<= "
		case "gte":
			condition.Main += ">= "
		case "in":
			condition.Main += "IN "
		case "like":
			condition.Main += "ILIKE "
		}

		condition.Main += "?"
		if i < len(filters)-1 {
			if isOr {
				condition.Main += " OR "
			} else {
				condition.Main += " AND "
			}
		}

		switch filter.Value.(type) {
		case bool:
			value := filter.Value.(bool)
			condition.Values = append(condition.Values, value)
		case string:
			value := strings.TrimSpace(filter.Value.(string))
			condition.Values = append(condition.Values, value)
		case []string:
			value := filter.Value.([]string)

			for i := range value {
				value[i] = strings.TrimSpace(value[i])
			}

			condition.Values = append(condition.Values, value)
		case int:
			value := strconv.Itoa(filter.Value.(int))
			condition.Values = append(condition.Values, value)
		case float64:
			value := fmt.Sprintf("%.0f", filter.Value)
			condition.Values = append(condition.Values, value)
		default:
			fmt.Println("not found", reflect.TypeOf(filter.Value))
		}
	}

	return
}
