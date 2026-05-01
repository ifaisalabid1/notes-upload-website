package storage

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
)

const (
	MaxUploadSize = 50 << 20
	sniffLen      = 512
)

var ErrFileTooLarge = fmt.Errorf("file exceeds maximum size of %d MB", MaxUploadSize>>20)

func DetectFileType(r io.Reader) (mimeType string, fileType domain.FileType, full io.Reader, err error) {
	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", "", nil, fmt.Errorf("read file header: %w", err)
	}

	mimeType = http.DetectContentType(buf[:n])

	ft, ok := domain.AllowedMIMETypes[mimeType]
	if !ok {
		return "", "", nil, fmt.Errorf(
			"unsupported file type %q: only PDF and images are allowed", mimeType,
		)
	}

	full = io.MultiReader(
		&byteReader{data: buf[:n]},
		r,
	)

	return mimeType, ft, full, nil
}

type byteReader struct {
	data []byte
}

func (b *byteReader) Read(p []byte) (int, error) {
	if len(b.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, b.data)
	b.data = b.data[n:]
	return n, nil
}

func ExtensionForMIME(mime string) string {
	switch mime {
	case "application/pdf":
		return "pdf"
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "bin"
	}
}
