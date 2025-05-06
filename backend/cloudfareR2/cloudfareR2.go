// r2_storage.go
package cloudfareR2

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// R2Storage handles file operations with Cloudflare R2
type R2Storage struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// NewR2Storage creates a new Cloudflare R2 client
func NewR2Storage(accessKey, secretKey, accountID, bucketName string) (*R2Storage, error) {
	// Create custom credentials provider
	credProvider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	// Load basic configuration with credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credProvider),
		config.WithRegion("auto"), // R2 uses "auto" as region
	)
	if err != nil {
		return nil, err
	}

	// Create an S3 client with a custom endpoint
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// This is the current recommended way to set endpoints - using BaseEndpoint at service client level
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))

		// If you need to use a custom endpoint
		// o.EndpointOptions.DisableHTTPS = false // set to true for HTTP
	})

	// Public URL pattern (adjust based on your bucket configuration)
	publicURL := os.Getenv("R2_PUBLIC_URL")

	return &R2Storage{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}, nil
}

// UploadFile uploads a file to Cloudflare R2
func (r *R2Storage) UploadFile(file multipart.File, fileHeader *multipart.FileHeader) (string, string, error) {
	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := uuid.New().String() + ext
	key := fmt.Sprintf("uploads/%s", filename)

	// Determine content type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to R2
	_, err := r.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", "", err
	}

	// Generate public URL for the uploaded file
	// Using R2 public URLs or custom domain if configured
	fileURL := fmt.Sprintf("%s/%s", r.publicURL, key)

	return filename, fileURL, nil
}

// DeleteFile deletes a file from Cloudflare R2
func (r *R2Storage) DeleteFile(filename string) error {
	key := fmt.Sprintf("uploads/%s", filename)

	_, err := r.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})

	return err
}

// GetPresignedURL generates a presigned URL for temporary direct access
func (r *R2Storage) GetPresignedURL(filename string, duration time.Duration) (string, error) {
	key := fmt.Sprintf("uploads/%s", filename)

	presignClient := s3.NewPresignClient(r.client)

	presignedURL, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(duration))

	if err != nil {
		return "", err
	}

	return presignedURL.URL, nil
}
