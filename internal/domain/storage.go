package domain

import (
	"context"
	"io"
)

type FileType string

const (
	FileTypePDF   FileType = "pdf"
	FileTypeImage FileType = "image"
)

var AllowedMIMETypes = map[string]FileType{
	"application/pdf": FileTypePDF,
	"image/jpeg":      FileTypeImage,
	"image/png":       FileTypeImage,
	"image/webp":      FileTypeImage,
}

type UploadInput struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
	FileType    FileType
}

type StorageService interface {
	Upload(ctx context.Context, input UploadInput) error
	Delete(ctx context.Context, key string) error
}
