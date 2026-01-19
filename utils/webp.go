package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"hyw-webpics/config"

	"github.com/google/uuid"
	"github.com/nickalie/go-webpbin"
)

// ConvertToWebP converts the uploaded image to WebP format using pre-compiled cwebp binary
func ConvertToWebP(file io.Reader, originalName string) (string, error) {
	// Generate unique filename
	filename := uuid.New().String() + ".webp"
	outputPath := filepath.Join(config.AppConfig.UploadDir, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(config.AppConfig.UploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create temporary file for input because go-webpbin works better with files
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Create WebP encoder
	// Quality 75 is a good balance
	encoder := webpbin.NewCWebP().
		Quality(75).
		InputFile(tempFile.Name()).
		OutputFile(outputPath)

	if err := encoder.Run(); err != nil {
		return "", fmt.Errorf("webp conversion failed: %w", err)
	}

	return filename, nil
}
