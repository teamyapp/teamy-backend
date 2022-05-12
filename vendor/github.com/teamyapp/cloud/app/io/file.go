package io

import (
	"fmt"
	"os"
	"path/filepath"
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
