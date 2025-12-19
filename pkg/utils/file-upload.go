package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// ValidateFileTypeFromContent extracts and validates file type based on actual file content.
// It reads the first 512 bytes of the file to detect the content type using http.DetectContentType.
// allowedTypes is a slice of allowed MIME types (e.g., []string{"image/jpeg", "image/png"}).
// Returns the validated content type and an error if validation fails.
func ValidateFileTypeFromContent(fileHeader *multipart.FileHeader, allowedTypes []string) (string, error) {
	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes for content type detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	if n == 0 {
		return "", fmt.Errorf("file is empty")
	}

	// Detect content type
	contentType := http.DetectContentType(buffer[:n])

	// Validate content type against allowed types
	allowedMap := make(map[string]bool, len(allowedTypes))
	for _, t := range allowedTypes {
		allowedMap[t] = true
	}

	if !allowedMap[contentType] {
		return "", fmt.Errorf("invalid file type: %s", contentType)
	}

	// Reset file pointer to beginning (file will be reopened when needed)
	// Note: The file handle is closed here, but gin will reopen it when SaveUploadedFile is called
	_, _ = file.Seek(0, 0)

	return contentType, nil
}

// getFileExtensionFromContentType returns the appropriate file extension (with leading dot)
// based on the detected content type. Returns empty string if content type is not recognized.
func GetFileExtensionFromContentType(contentType string) string {
	extensionMap := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}

	if ext, ok := extensionMap[contentType]; ok {
		return ext
	}
	return ""
}
