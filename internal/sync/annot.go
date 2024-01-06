package sync

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/logger"
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	annotPath = "/Annot"
)

func loadAnnotFiles(syncData model.Sync, zipLocation, saveLocation string) (newestTime int64, err error) {
	logger.Log.Info("started annot load files")

	if !syncData.IsFullSynced {
		syncData.LastAnnotSyncDate, err = loadAnnotsFull(zipLocation, saveLocation)
		if err != nil {
			logger.Log.Error("failed to load annots",
				zap.Error(err),
			)

			return 0, err
		}
	}

	logger.Log.Info("started partial annot load files")

	syncData.LastAnnotSyncDate, err = loadAnnotsPartial(zipLocation, saveLocation, syncData)
	if err != nil {
		logger.Log.Error("failed to load partial annots",
			zap.Error(err),
		)
		return 0, err
	}

	return syncData.LastAnnotSyncDate, nil
}

func loadAnnotsFull(zipLocation, saveLocation string) (newestTime int64, err error) {
	logger.Log.Info("started annots full sync")
	// FTP server URL
	ftpClient, err := ConnectToFtp(ftpUrl)
	if err != nil {
		logger.Log.Error("failed to connect to ftp",
			zap.Error(err),
		)
		return
	}
	defer ftpClient.Close()

	rootPath := annotPath

	entries, err := ftpClient.ReadDir(rootPath)
	if err != nil {
		logger.Log.Error("failed to list annot entries in ftp",
			zap.Error(err),
		)

		return
	}

	for _, entry := range entries {
		splits := strings.Split(entry.Name(), "_")
		if len(splits) >= 2 {
			continue
		}

		directoryPath := filepath.Join(rootPath, entry.Name())
		entriesMonths, err := ftpClient.ReadDir(directoryPath)
		if err != nil {
			return 0, err
		}

		for _, entriesMonth := range entriesMonths {
			zipPath := filepath.Join(directoryPath, entriesMonth.Name())
			entriesAnnots, err := ftpClient.ReadDir(zipPath)
			if err != nil {
				return 0, err
			}

			for _, entriesAnnot := range entriesAnnots {
				annotPath := filepath.Join(zipPath, entriesAnnot.Name())

				splits := strings.Split(entriesAnnot.Name(), "_")
				if len(splits) < 2 {
					continue
				}

				layout := "20060102"
				t, err := time.Parse(layout, splits[0])
				if err != nil {
					return 0, err
				}

				if t.Unix() > newestTime {
					newestTime = t.Unix()
				}

				err = DownloadAndDecompress(ftpClient, annotPath, zipLocation, saveLocation)
				if err != nil {
					return 0, err
				}

				err = UploadAnnots(saveLocation)
				if err != nil {
					return 0, err
				}
			}
		}
	}

	return
}

func loadAnnotsPartial(zipLocation, saveLocation string, syncData model.Sync) (newestTime int64, err error) {
	logger.Log.Info("started annots partial sync",
		zap.Int64("lastSyncDate", syncData.LastAnnotSyncDate),
	)

	// FTP server URL
	ftpClient, err := ConnectToFtp(ftpUrl)
	if err != nil {
		logger.Log.Error("failed to connect to ftp",
			zap.Error(err),
		)
		return
	}
	defer ftpClient.Close()

	rootPath := annotPath
	entries, err := ftpClient.ReadDir(rootPath)
	if err != nil {
		logger.Log.Error("failed to list annot entries in ftp",
			zap.Error(err),
		)

		return
	}

	for _, entry := range entries {
		splits := strings.Split(entry.Name(), "_")
		if len(splits) < 2 {
			continue
		}

		layout := "20060102"
		t, err := time.Parse(layout, splits[0])
		if err != nil {
			return 0, err
		}

		if t.Unix() <= syncData.LastAnnotSyncDate {
			continue
		}

		if t.Unix() > newestTime {
			newestTime = t.Unix()
		}

		filePath := filepath.Join(rootPath, entry.Name())
		err = DownloadAndDecompress(ftpClient, filePath, zipLocation, saveLocation)
		if err != nil {
			return 0, err
		}

		err = UploadAnnots(saveLocation)
		if err != nil {
			return 0, err
		}
	}

	return
}

func UploadAnnots(saveLocation string) (err error) {
	defer os.RemoveAll(saveLocation)

	annotFiles, err := os.ReadDir(saveLocation)
	if err != nil {
		return err
	}

	var (
		productUpdate = make([]model.Product, 0)
		skipped       = 0
		descCount     = 0
		imageCount    = 0
		count         = 0
	)

	for _, annotFile := range annotFiles {
		count++
		if count%1000 == 0 {
			time.Sleep(time.Millisecond * 50)
		}

		annotType := annotFile.Name()[5:8]
		id := strings.TrimLeft(strings.Split(annotFile.Name()[10:], ".")[0], "0")

		var product model.Product
		if err := database.DB.Where("id = ?", id).First(&product).Error; err != nil {
			skipped++
			continue
		}

		filePath := filepath.Join(saveLocation, annotFile.Name())
		switch annotType {
		case "ATX":
			raw, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			descCount++

			re := regexp.MustCompile("<[^>]*>")
			cleanText := re.ReplaceAllString(string(raw), "")

			product.Description = strings.TrimSpace(cleanText)
		case "COP":
			raw, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			product.CoverPicture = base64.StdEncoding.EncodeToString(raw)
			imageCount++
		default:
			continue
		}

		productUpdate = append(productUpdate, product)
	}

	logger.Log.Debug("starting annot upload",
		zap.Int("skipped", skipped),
		zap.Int("adding", len(productUpdate)),
		zap.Int("imageCount", imageCount),
		zap.Int("descCount", descCount),
	)

	wg := sync.WaitGroup{}
	wg.Add(2)
	err = batchUpdate(productUpdate[0:len(productUpdate)/2], &wg)
	if err != nil {
		return err
	}

	err = batchUpdate(productUpdate[len(productUpdate)/2:], &wg)
	if err != nil {
		return err
	}
	wg.Wait()
	return
}
