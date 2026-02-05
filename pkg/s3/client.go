package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "github.com/Alkush-Pipania/Scrapper/config"
)

type Client struct {
	s3Client   *s3.Client
	downloader *manager.Downloader
	uploader   *manager.Uploader
	bucketName string
	endpoint   string
}

func NewClient(ctx context.Context, cfg appconfig.ClientConfig) (*Client, error) {
	baseEndpoint, publicHost, err := normalizeEndpoint(cfg.Endpoint, cfg.BucketName)
	if err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(baseEndpoint)
		o.UsePathStyle = false // Use virtual-hosted style (bucket.endpoint) for DigitalOcean Spaces
	})

	return &Client{
		s3Client:   s3Client,
		downloader: manager.NewDownloader(s3Client),
		uploader:   manager.NewUploader(s3Client),
		bucketName: cfg.BucketName,
		endpoint:   publicHost,
	}, nil
}

func normalizeEndpoint(endpoint string, bucket string) (baseEndpoint string, publicHost string, err error) {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		return "", "", fmt.Errorf("s3 endpoint is empty")
	}

	// Ensure scheme for parsing / SDK base endpoint
	withScheme := trimmed
	if !strings.Contains(trimmed, "://") {
		withScheme = "https://" + trimmed
	}

	parsed, err := url.Parse(withScheme)
	if err != nil {
		return "", "", fmt.Errorf("invalid s3 endpoint: %w", err)
	}

	host := parsed.Host
	if host == "" {
		return "", "", fmt.Errorf("invalid s3 endpoint host: %s", endpoint)
	}

	// If user included bucket in endpoint, strip it for host/public URL
	if bucket != "" && strings.HasPrefix(host, bucket+".") {
		host = strings.TrimPrefix(host, bucket+".")
	}

	return parsed.Scheme + "://" + host, host, nil
}
