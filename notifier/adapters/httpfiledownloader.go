package adapters

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/4Kaze/birthdaybot/common"
)

type HttpFileDownloader struct{}

func NewHttpFileDownloader() *HttpFileDownloader {
	return &HttpFileDownloader{}
}

func (HttpFileDownloader) Download(ctx context.Context, link string) (string, error) {
	tmpFile, err := os.CreateTemp("", "*")
	if err != nil {
		common.ErrorLogger.Printf("Failed to create temp file to download image: %s due to: %v\n", link, err)
		return "", err
	}
	defer tmpFile.Close()
	response, err := http.Get(link)
	if err != nil {
		common.ErrorLogger.Printf("Failed to get image from url: %s due to: %v\n", link, err)
		return "", err
	}
	defer response.Body.Close()
	_, err = io.Copy(tmpFile, response.Body)
	if err != nil {
		common.ErrorLogger.Printf("Failed to copy image contents from url: %s to file: %s due to: %v\n", link, tmpFile.Name(), err)
		return "", err
	}
	filePath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		common.ErrorLogger.Printf("Failed to obtain a file path from: %s due to: %v\n", tmpFile.Name(), err)
		return "", err
	}
	return filePath, nil
}
