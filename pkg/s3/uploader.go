package s3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (c *Client) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to s3: %w", err)
	}

	return fmt.Sprintf("https://%s.%s/%s", c.bucketName, c.endpoint, key), nil
}
