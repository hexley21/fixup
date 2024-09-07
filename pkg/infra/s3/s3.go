package s3

import (
	"context"
	"io"
)

type Bucket interface {
	PutObject(ctx context.Context, file io.Reader, directory string, fileName string, fileSize int64, contentType string) (string, error)
	GetObject(ctx context.Context, fileName string) (io.Reader, error)
	DeleteObject(ctx context.Context, fileName string) error
}