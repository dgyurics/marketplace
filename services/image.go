package services

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	ImageUploadPath = "images"
)

type ImageService interface {
	StoreImage(productID string, file io.Reader, filename string) (string, error)
	IsSupportedImage(file io.Reader) (bool, error)
}

type imageService struct{}

func NewImageService() ImageService {
	return &imageService{}
}

func (s *imageService) StoreImage(productID string, file io.Reader, filename string) (string, error) {
	// Create subdirectory for the file
	dirPath := filepath.Join(ImageUploadPath, productID)
	filePath := filepath.Join(dirPath, filepath.Base(filename))
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", err
	}

	// Save the file
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

func (s *imageService) IsSupportedImage(file io.Reader) (bool, error) {
	// Read the first 512 bytes to check the file type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		return false, err
	}

	// Reset the read pointer
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return false, err
		}
	}

	contentType := http.DetectContentType(buff)
	return contentType == "image/jpeg" || contentType == "image/png", nil
}
