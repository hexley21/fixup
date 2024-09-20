package cdn

import "context"

type FileInvalidator interface {
	InvalidateFile(ctx context.Context, fileName string) error
}
