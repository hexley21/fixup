package s3

import (
	"context"
	"io"

	aws_cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hexley21/handy/pkg/config"
)

type awsS3 struct {
	s3Client *s3.Client
	bucket   string
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
	}

	return s3, nil
}

func (as3 *awsS3) PutObject(ctx context.Context, file io.Reader, fileName string, fileSize int64, fileType string) error {
	_, err := as3.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &as3.bucket,
		Body:          file,
		Key:           &fileName,
		ContentLength: &fileSize,
		ContentType:   &fileType,
	})

	return err
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
