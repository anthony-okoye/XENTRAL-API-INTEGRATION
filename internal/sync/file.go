package sync

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jlaffaye/ftp"
)

func RetrieveAndSave(ftpClient *ftp.ServerConn, inputPath, outputPath string) (err error) {
	res, err := ftpClient.Retr(inputPath)
	if err != nil {
		return err
	}
	defer res.Close()

	dirPath := filepath.Dir(outputPath)
	os.MkdirAll(dirPath, 0744)

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res)
	if err != nil {
		return err
	}

	return
}

func DecompressToDisk(zipLocation string, resultLocation string) error {
	// Open the zip file
	cmd := exec.Command("unzip", "-o", zipLocation, "-d", resultLocation)
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}
