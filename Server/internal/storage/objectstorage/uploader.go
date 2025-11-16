package objectstorage

import (
	"context"
	"fmt"
	"io"

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


