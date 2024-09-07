package s3

import (
	"context"
	"io"
)

type S3 interface {
	PutObject(ctx context.Context, file io.Reader, fileName string, fileSize int64, fileType string) error
	GetObject(ctx context.Context, fileName string) (io.Reader, error)
	DeleteObject(ctx context.Context, fileName string) error
}