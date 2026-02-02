package s3

import (
	"context"

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
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = false // Use virtual-hosted style (bucket.endpoint) for DigitalOcean Spaces
	})

	return &Client{
		s3Client:   s3Client,
		downloader: manager.NewDownloader(s3Client),
		uploader:   manager.NewUploader(s3Client),
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
	}, nil
}
