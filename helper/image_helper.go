package helper

import (
	"io"
	"os"
	"path/filepath"
)

func SaveImage(file io.Reader, filename, folderPath string) (string, error) {
	filePath := filepath.Join(folderPath, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return filePath, nil
}


func DeleteImage(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}
