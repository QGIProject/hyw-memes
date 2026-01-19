package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"hyw-webpics/config"

	"github.com/google/uuid"
)

// ConvertToWebP converts the uploaded image to WebP format using system cwebp
func ConvertToWebP(file io.Reader, originalName string) (string, error) {
	// Generate unique filename
	filename := uuid.New().String() + ".webp"
	outputPath := filepath.Join(config.AppConfig.UploadDir, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(config.AppConfig.UploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create temporary file for input
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Create WebP encoder using system cwebp
	// Quality 75 is a good balance
	cmd := exec.Command("cwebp", "-q", "75", tempFile.Name(), "-o", outputPath)
	if err := cmd.Run(); err != nil {
		// Fallback: If cwebp is not in path, it might be in .bin (for local dev)
		localCwebp := filepath.Join(".", ".bin", "webp", "cwebp.exe")
		cmd = exec.Command(localCwebp, "-q", "75", tempFile.Name(), "-o", outputPath)
		if err2 := cmd.Run(); err2 != nil {
			return "", fmt.Errorf("webp conversion failed (system cwebp err: %v, local cwebp err: %v)", err, err2)
		}
	}

	return filename, nil
}
