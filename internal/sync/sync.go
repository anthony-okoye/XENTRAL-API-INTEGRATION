package sync

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"go.uber.org/zap"
)

const (
	onixZipLocation   = "/tmp/onix/bz/zip/data.zip"
	onixDataLocation  = "/tmp/onix/bz/data"
	annotZipLocation  = "/tmp/onix/annot/zip/data.zip"
	annotDataLocation = "/tmp/onix/annot/data"

	annotLocation = "onix/annot"

	ftpUrl = "sftp.buchzentrum.ch:22"
)

var (
	dayToSync = time.Now()
)

func Worker() {
	err := SeedCategories(logger.Log)
	if err != nil {
		logger.Log.Error("failed to sync",
			zap.Error(err),
		)
		return
	}

	for {
		start := time.Now()

		if start.After(dayToSync) {
			logger.Log.Info("starting sync",
				zap.String("date", dayToSync.String()),
			)

			err = SyncData()
			if err != nil {
				logger.Log.Error("failed to sync, will try again",
					zap.Error(err),
				)

				time.Sleep(time.Second * 60)
				continue
			}

			dayToSync = dayToSync.AddDate(0, 0, 1)
			fmt.Println(start)
			fmt.Println(time.Now())
		}

		time.Sleep(time.Second * 60)
	}

}

var count int

func SyncData() (err error) {
	logger.Log.Info("started sync")

	var syncData model.Sync
	res := database.DB.Where("id = ?", "1").First(&syncData)
	if res.RowsAffected == 0 {
		logger.Log.Warn("failed to load sync data from the database", zap.Error(err))
		return
	}

	date2, err := loadOnixFiles(syncData, onixZipLocation, onixDataLocation, false)
	if err != nil {
		logger.Log.Error("failed to load onix files",
			zap.Error(err),
		)
		return
	}

	err = loadOnixData(onixDataLocation, false)
	if err != nil {
		logger.Log.Error("failed to sync data",
			zap.Error(err),
		)
		return
	}

	date1, err := loadOnixFiles(syncData, onixZipLocation, onixDataLocation, true)
	if err != nil {
		logger.Log.Error("failed to load onix files",
			zap.Error(err),
		)
		return
	}

	err = loadOnixData(onixDataLocation, true)
	if err != nil {
		logger.Log.Error("failed to sync data",
			zap.Error(err),
		)
		return
	}

	_, err = loadAnnotFiles(syncData, annotZipLocation, annotDataLocation)
	if err != nil {
		logger.Log.Error("failed to load annot files",
			zap.Error(err),
		)
		return
	}

	// take the older timestamp
	syncData.LastOnixSyncDate = date1
	if date1 == 0 && date2 == 0 {
		return
	} else if date1 == 0 {
		syncData.LastOnixSyncDate = date2
	} else if date2 == 0 {
		syncData.LastOnixSyncDate = date1
	} else if date1 > date2 {
		syncData.LastOnixSyncDate = date2
	}

	syncData.LastAnnotSyncDate = syncData.LastOnixSyncDate
	syncData.IsFullSynced = true
	err = database.DB.Updates(&syncData).Error
	if err != nil {
		logger.Log.Error("failed to update annotations", zap.Error(err))
		return
	}

	return
}

func DownloadAndDecompress(ftpClient *sftp.Client, ftpLocation, zipLocation, dataLocation string) (err error) {
	res, err := ftpClient.Open(ftpLocation)
	if err != nil {
		return err
	}
	defer res.Close()

	directoryPath := filepath.Dir(zipLocation)
	os.MkdirAll(directoryPath, 0744)

	file, err := os.Create(zipLocation)
	if err != nil {
		return err
	}
	defer file.Close()

	logger.Log.Info("copying zip file to disk",
		zap.String("ftpLocation", ftpLocation),
		zap.String("zipLocation", zipLocation),
		zap.String("dataLocation", dataLocation),
	)

	_, err = io.Copy(file, res)
	if err != nil {
		return err
	}

	logger.Log.Info("decompressing zip file to disk",
		zap.String("ftpLocation", ftpLocation),
		zap.String("zipLocation", zipLocation),
		zap.String("dataLocation", dataLocation),
	)

	err = DecompressToDisk(zipLocation, dataLocation)
	if err != nil {
		return err
	}

	return
}
