package s3

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hexley21/handy/pkg/config"
)

type awsS3 struct {
	s3Client *s3.Client
	bucket   string
	nameSize int
}

func NewClient(awsCfg config.AWSCfg, s3Cfg config.S3) (*awsS3, error) {
	clientCfg, err := awsCfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	s3 := &awsS3{
		s3Client: s3.NewFromConfig(clientCfg),
		bucket:   s3Cfg.Bucket,
		nameSize: s3Cfg.RandomNameSize,
	}

	return s3, nil
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (as3 *awsS3) PutObject(ctx context.Context, file io.Reader, directory string, fileName string, fileSize int64, fileType string) (string, error) {
	if fileName == "" {
		randomString, err := generateRandomString(as3.nameSize)
		if err != nil {
			return "", err
		}
		fileName = randomString
	}

	var keyBuilder strings.Builder
	keyBuilder.WriteString(directory)
	keyBuilder.WriteString(fileName)

	key := keyBuilder.String()

	_, err := as3.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &as3.bucket,
		Body:          file,
		Key:           &key,
		ContentLength: &fileSize,
		ContentType:   &fileType,
	})

	return fileName, err
}

func (as3 *awsS3) GetObject(ctx context.Context, file string) (io.Reader, error) {
	output, err := as3.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &as3.bucket,
		Key:    &file,
	})

	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

func (as3 *awsS3) DeleteObject(ctx context.Context, file string) error {
	_, err := as3.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &as3.bucket,
		Key:    &file,
	})

	return err
}
