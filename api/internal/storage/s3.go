package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "github.com/ddddami/laivan/internal/config"
)

type UploadInput struct {
	Key         string
	Body        io.Reader
	ContentType string
}

type Uploader interface {
	Upload(ctx context.Context, input UploadInput) (string, error)
}

type S3Uploader struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

func NewS3Uploader(ctx context.Context, cfg appconfig.MediaConfig) (*S3Uploader, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		o.UsePathStyle = true
	})

	return &S3Uploader{
		client:        client,
		bucket:        cfg.S3Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, input UploadInput) (string, error) {
	if strings.TrimSpace(input.Key) == "" {
		return "", fmt.Errorf("object key is required")
	}
	if input.Body == nil {
		return "", fmt.Errorf("upload body is required")
	}

	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(input.Key),
		Body:        input.Body,
		ContentType: aws.String(input.ContentType),
	})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return u.publicBaseURL + "/" + strings.TrimLeft(input.Key, "/"), nil
}

func MediaObjectKey(targetType, targetID, mediaID, filename string) string {
	return path.Join("media", safePathPart(targetType), safePathPart(targetID), safePathPart(mediaID), safeFilename(filename))
}

func safePathPart(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

func safeFilename(filename string) string {
	filename = path.Base(strings.TrimSpace(filename))
	if filename == "." || filename == "/" || filename == "" {
		return "upload"
	}

	var b strings.Builder
	for _, r := range filename {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "upload"
	}
	return b.String()
}
