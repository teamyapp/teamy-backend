package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func CreateFile(filePath string) error {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	defer file.Close()
	return err
}

func CreateFileWithLog(filePath string) error {
	err := CreateFile(filePath)
	if err == nil {
		fmt.Printf("File created at: %s\n", filePath)
	}
	return err
}

func GetFileURL(cloudWebAPIBaseURL string, fileID uint64) string {
	fileIDParam := strconv.FormatUint(fileID, 10)
	return fmt.Sprintf("%s/file/files/%s", cloudWebAPIBaseURL, fileIDParam)
}
