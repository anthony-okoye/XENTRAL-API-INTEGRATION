package model

import (
	"bookbox-backend/pkg/logger"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type Root struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Active    *bool     `json:"active" gorm:"column:active"`
	CreatedAt time.Time `json:"created_at" gorm:"<-:create"`
	UpdatedAt time.Time `json:"-" gorm:"index"`
}

const (
	imageMBLimit = 10
)

func UploadImage(rootPath, id string, imageB64 string) (filePath string, err error) {
	// Check the size of the image data
	imageSizeInBytes := len(imageB64)
	imageSizeInMB := float64(imageSizeInBytes) / (1024 * 1024)

	logger.Log.Debug("image uploading",
		zap.String("id", id),
		zap.Float64("imageSizeInMB", imageSizeInMB),
	)

	if imageSizeInMB > imageMBLimit {
		err = fmt.Errorf("image uploaded cannot be more than 2mb")
		log.Fatal("OVER LIMIT")
		return
	}

	subName := "000"
	if len(id) >= 3 {
		subName = id[0:3]
	}
	subFolder := filepath.Join(rootPath, subName)
	os.MkdirAll(subFolder, 0744)
	filePath = filepath.Join(subFolder, id)
	raw, err := base64.StdEncoding.DecodeString(imageB64)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filePath, raw, 0744)
	if err != nil {
		return "", err
	}

	return
}
