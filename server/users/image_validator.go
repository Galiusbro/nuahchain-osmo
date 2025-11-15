package users

import (
	"errors"
	"io"
	"mime/multipart"
)

// ImageValidator validates uploaded images
// Currently accepts all images - validation will be added later
type ImageValidator struct {
	// Future: max file size, allowed formats, etc.
}

// NewImageValidator creates a new image validator
func NewImageValidator() *ImageValidator {
	return &ImageValidator{}
}

// ValidateImage validates an uploaded image file
// Currently always accepts - validation will be added later
func (v *ImageValidator) ValidateImage(file multipart.File, header *multipart.FileHeader) error {
	// TODO: Add validation for:
	// - File size limits
	// - Image format (JPEG, PNG, WebP, etc.)
	// - Image dimensions
	// - Content scanning for inappropriate content
	// - Virus scanning

	// For now, just check that file exists
	if file == nil {
		return errors.New("file is required")
	}

	if header == nil {
		return errors.New("file header is required")
	}

	// Check file size (basic check - max 10MB for now)
	if header.Size > 10*1024*1024 {
		return errors.New("file size exceeds 10MB limit")
	}

	// Check that we can read at least some bytes
	buf := make([]byte, 1)
	_, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return errors.New("invalid file")
	}

	// Reset file pointer for actual upload
	file.Seek(0, 0)

	return nil
}

// GetAllowedFormats returns list of allowed image formats
func (v *ImageValidator) GetAllowedFormats() []string {
	return []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
		"image/gif",
	}
}

// GetMaxFileSize returns maximum allowed file size in bytes
func (v *ImageValidator) GetMaxFileSize() int64 {
	return 10 * 1024 * 1024 // 10MB
}
