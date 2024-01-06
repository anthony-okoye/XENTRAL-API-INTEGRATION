package crud

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/execute/postrun"
	"bookbox-backend/internal/execute/prerun"
	"bookbox-backend/internal/query"
	"bookbox-backend/internal/request"
	"bookbox-backend/internal/route/auth"
	"bookbox-backend/internal/route/fail"
	"bookbox-backend/internal/server/router"
	"bookbox-backend/pkg/logger"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func UpdateHandler(ctx *gin.Context) {
	var (
		updateRequest  = request.Request{}
		updateResponse = request.Response{}
	)

	// bind input data to request format
	err := ctx.ShouldBindJSON(&updateRequest)
	if err != nil {
		logger.Log.Error("Failed to bind input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, logger.Log)
		return
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.String("entity", updateRequest.Entity),
	))

	log.Info("update started")

	issuer, err := auth.GetIssuer(ctx)
	if err != nil {
		errMsg := "authentication failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user auth is incorrect")
		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 403, log)
		return
	}

	if updateRequest.Entity == "order" || updateRequest.Entity == "product" {
		productMutex.Lock()
		defer productMutex.Unlock()
	}

	row := query.Determine(updateRequest.Entity)
	if row == nil {
		err = fmt.Errorf("entity does not exist")
		log.Error("Failed to parse input data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	// check authorisation
	if !IsAuthorized(issuer, "update", updateRequest.Entity) {
		errMsg := "authorization failed"
		log.Error(errMsg,
			zap.Error(err),
		)

		err = fmt.Errorf("user is not authorized for this request")
		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 403, log)
		return
	}

	// run prerun functions if they exist
	if f, exist := prerun.UpdateFunctions[updateRequest.Entity]; exist {
		err = f(&updateRequest, issuer)
		if err != nil {
			log.Error("failed to run prerun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	raw, err := json.Marshal(updateRequest.Data)
	if err != nil {
		log.Error("Failed to marshal data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	id, ok := updateRequest.Data["id"].(string)
	if !ok {
		err = fmt.Errorf("id is not specified")
		log.Error("Incorrect format in input params",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	err = json.Unmarshal(raw, &row)
	if err != nil {
		log.Error("Failed to unmarshal data",
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	err = UpdateTransaction(updateRequest, database.DB, row, id)
	if err != nil {
		log.Warn("Update failed",
			zap.String("id", id),
			zap.Error(err),
		)

		fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
		return
	}

	// run postrun functions if they exist
	if f, exist := postrun.UpdateFunctions[updateRequest.Entity]; exist {
		err = f(&updateRequest, issuer, log)
		if err != nil {
			log.Error("failed to run postrun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	log.Info("update finished")

	updateResponse.Status = true

	// run postrun cache functions if they exist
	if f, exist := postrun.CacheUpdateFunctions[updateRequest.Entity]; exist {
		err = f(updateRequest, updateResponse, issuer, log)
		if err != nil {
			log.Error("failed to run postrun function",
				zap.Error(err),
			)

			fail.ReturnError(ctx, updateResponse, []string{err.Error()}, 400, log)
			return
		}
	}

	ctx.JSON(200, updateResponse)
}

var (
	dataCart = map[string]string{
		"cart_items": "cart_items",
	}
	dataCategory = map[string]string{
		"products": "product_categories",
	}
	dataProduct = map[string]string{
		"categories":     "product_categories",
		"sales_channels": "sales_channel_products",
	}

	dataOrder = map[string]string{
		"products": "order_items",
	}

	dataUser = map[string]string{
		"favorites": "favorites",
	}
)

func Mapper(entity, val string) (string, error) {
	var ok bool

	switch entity {
	case "cart":
		if val, ok = dataCart[val]; ok {
			return val, nil
		}
	case "category":
		if val, ok = dataCategory[val]; ok {
			return val, nil
		}
	case "product":
		if val, ok = dataProduct[val]; ok {
			return val, nil
		}
	case "user":
		if val, ok = dataUser[val]; ok {
			return val, nil
		}
	case "order":
		if val, ok := dataOrder[val]; ok {
			return val, nil
		}

	default:
		return "", fmt.Errorf("entity not supported for override to update")
	}

	return "", fmt.Errorf("override to update does not exist")
}

func UpdateTransaction(updateRequest request.Request, db *gorm.DB, row any, id string) (err error) {
	tx := database.DB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	var count int64
	tx.Model(row).Where("id = ?", id).Count(&count)
	if count == 0 {
		return fmt.Errorf("entity with specified id does not exist")
	}

	// remove previous relationships of specified entities
	entityRow := query.Determine(updateRequest.Entity)
	raw, _ := json.Marshal(updateRequest.Data)
	json.Unmarshal(raw, &entityRow)

	for _, val := range updateRequest.Metadata.OverrideOnUpdate {
		val, err = Mapper(updateRequest.Entity, val)
		if err != nil {
			return err
		}

		// Delete all the associated cart items from the "cart_items" table
		query := fmt.Sprintf("DELETE FROM %s WHERE %s_id = ?", val, strings.ToLower(updateRequest.Entity))
		if err = tx.Exec(query, id).Error; err != nil {
			return err
		}
	}

	if len(updateRequest.Metadata.UpdateFields) != 0 {
		// add updated version
		err = tx.Select(updateRequest.Metadata.UpdateFields).Updates(row).Error
		if err != nil {
			return
		}
	} else {
		// add updated version
		err = tx.Select("*").Updates(row).Error
		if err != nil {
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		return
	}

	return nil
}

func init() {
	router.Router.Handle("POST", "/update", UpdateHandler)
}
