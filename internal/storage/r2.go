package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ifaisalabid1/notes-upload-website/internal/config"
	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
)

type r2Storage struct {
	client     *s3.Client
	bucketName string
}

func NewR2Storage(cfg config.R2Config) (domain.StorageService, error) {
	r2Endpoint := fmt.Sprintf(
		"https://%s.r2.cloudflarestorage.com",
		cfg.AccountID,
	)

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(r2Endpoint),
		Region:       "auto",
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		),
	})

	slog.Info("R2 storage initialized", "bucket", cfg.BucketName)

	return &r2Storage{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (s *r2Storage) Upload(ctx context.Context, input domain.UploadInput) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(s.bucketName),
		Key:                aws.String(input.Key),
		Body:               input.Body,
		ContentType:        aws.String(input.ContentType),
		ContentLength:      aws.Int64(input.Size),
		ContentDisposition: aws.String("inline"),
		CacheControl:       aws.String("private, no-store"),
	})
	if err != nil {
		return fmt.Errorf("r2 put object: %w", err)
	}

	slog.Info("file uploaded to R2", "key", input.Key, "size", input.Size)
	return nil
}

func (s *r2Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("r2 delete object: %w", err)
	}

	slog.Info("file deleted from R2", "key", key)
	return nil
}
