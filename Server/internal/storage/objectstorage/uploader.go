package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Uploader struct {
	client   *minio.Client
	bucket   string
	cdnBase  string
}

func NewUploader(endpoint, accessKey, secretKey, bucket, cdnBase string, secure bool) (*Uploader, error) {
	cl, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, err
	}
	return &Uploader{client: cl, bucket: bucket, cdnBase: cdnBase}, nil
}

func (u *Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := u.client.PutObject(ctx, u.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	if u.cdnBase != "" {
		return fmt.Sprintf("%s/%s", u.cdnBase, key), nil
	}
	// fallback: use virtual-hosted-style URL
	return fmt.Sprintf("https://%s.%s/%s", u.bucket, u.client.EndpointURL().Host, key), nil
}

// GetPublicURL generates a public URL for a stored file.
// If CDN base URL is configured, it uses that. Otherwise, it generates a presigned URL.
func (u *Uploader) GetPublicURL(ctx context.Context, storedURL string, expiry time.Duration) (string, error) {
	// If CDN base is configured and URL already starts with it, return as-is
	if u.cdnBase != "" && strings.HasPrefix(storedURL, u.cdnBase) {
		return storedURL, nil
	}

	// Extract key from stored URL
	key := u.extractKeyFromURL(storedURL)
	if key == "" {
		return storedURL, nil // Return original if we can't extract key
	}

	// If CDN base is configured, use it
	if u.cdnBase != "" {
		return fmt.Sprintf("%s/%s", u.cdnBase, key), nil
	}

	// Otherwise, generate a presigned URL
	presignedURL, err := u.client.PresignedGetObject(ctx, u.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// extractKeyFromURL extracts the object key from a storage URL.
// Handles formats like:
// - https://bucket.endpoint.com/key
// - https://endpoint.com/bucket/key
// - https://cdn.example.com/key
func (u *Uploader) extractKeyFromURL(storedURL string) string {
	parsed, err := url.Parse(storedURL)
	if err != nil {
		return ""
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	
	// If path starts with bucket name, remove it
	if strings.HasPrefix(path, u.bucket+"/") {
		path = strings.TrimPrefix(path, u.bucket+"/")
	}
	
	return path
}


