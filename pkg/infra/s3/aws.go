package s3

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/hexley21/handy/pkg/config"

	aws_cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsS3 struct {
	s3Client *s3.Client
	bucket   string
	nameSize int
}

func NewClient(cfg config.S3) (*awsS3, error) {
	s3Cfg, err := aws_cfg.LoadDefaultConfig(context.TODO(),
		aws_cfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		aws_cfg.WithRegion(cfg.Region),
	)

	if err != nil {
		return nil, err
	}

	s3 := &awsS3{
		s3Client: s3.NewFromConfig(s3Cfg),
		bucket:   cfg.Bucket,
		nameSize: cfg.RandomNameSize,
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

func (as3 *awsS3) PutObject(ctx context.Context, file io.Reader, fileName string, fileSize int64, fileType string) (string, error) {
	if fileName == "" {
		randomString, err := generateRandomString(as3.nameSize)
		if err != nil {
			return "", err
		}
		fileName = randomString
	}

	_, err := as3.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &as3.bucket,
		Body:          file,
		Key:           &fileName,
		ContentLength: &fileSize,
		ContentType:   &fileType,
	})

	return fileName, err
}

func (as3 *awsS3) GetObject(ctx context.Context, fileName string) (io.Reader, error) {
	output, err := as3.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &as3.bucket,
		Key:    &fileName,
	})

	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

func (as3 *awsS3) DeleteObject(ctx context.Context, fileName string) error {
	_, err := as3.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &as3.bucket,
		Key:    &fileName,
	})

	return err
}
