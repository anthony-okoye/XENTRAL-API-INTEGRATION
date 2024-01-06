package sync

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"embed"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

var (
	//go:embed categories.xlsx
	content embed.FS
)

func SeedCategories(log *zap.Logger) (err error) {
	f, err := content.Open("categories.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	xl, err := excelize.OpenReader(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := xl.GetRows("WGR")
	if err != nil {
		fmt.Println(err)
		return
	}

	var salesChannel model.SalesChannel
	res := database.DB.Preload("Categories").Where("id = ?", "1").First(&salesChannel)
	if res.RowsAffected == 0 {
		log.Fatal("failed to start saleschannel doesn't exist")
		return
	}

	if len(salesChannel.Categories) > 0 {
		log.Info("categories already loaded in", zap.Error(err))
		return
	}

	var parentIDStack []string = make([]string, 0)

	for i := 1; i < len(rows)-1; i++ {
		splits := strings.Split(rows[i][1], "^")
		nextSplits := strings.Split(rows[i+1][1], "^")

		prevID := ""
		if len(parentIDStack) != 0 {
			prevID = parentIDStack[len(parentIDStack)-1]
		}

		if len(splits) < len(nextSplits) {
			parentIDStack = append(parentIDStack, rows[i][0])
		} else if len(splits) > len(nextSplits) {
			parentIDStack = parentIDStack[:len(parentIDStack)-(len(splits)-len(nextSplits))]
		}

		if len(splits) <= 1 {
			url := ""
			if len(rows[i]) > 2 {
				url = rows[i][2]
			}
			rootCat := model.Category{
				Name: splits[0],
				URL:  url,
				Root: model.Root{
					ID: rows[i][0],
				},
			}

			salesChannel.Categories = append(salesChannel.Categories, rootCat)

			err = database.DB.Updates(&salesChannel).Error
			if err != nil {
				log.Info("categories already loaded in", zap.Error(err))
				return
			}
			continue
		}

		newCategory := splits[len(splits)-1]

		if newCategory == "" {
			log.Error("failed to load categories, name empty")
			return
		}
		var categories []model.Category

		res = database.DB.Preload("SubCategories").Find(&categories, "id = ?", prevID)

		if res.Error != nil {
			log.Error("failed to load categories",
				zap.Error(err),
			)

			return
		}

		url := ""
		if len(rows[i]) > 2 {
			url = rows[i][2]
		}

		var new = model.Category{
			Name: newCategory,
			URL:  url,
			Root: model.Root{
				ID: rows[i][0],
			},
		}

		for _, prevCategory := range categories {
			prevCategory.SubCategories = append(prevCategory.SubCategories, &new)

			err = database.DB.Updates(&prevCategory).Error
			if err != nil {
				log.Error("failed to update categories",
					zap.Error(err),
				)

				return err
			}
		}
	}

	return
}
