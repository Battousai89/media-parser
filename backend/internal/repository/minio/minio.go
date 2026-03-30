package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Bucket   string
	UseSSL   bool
}

type Client struct {
	config *Config
	client *minio.Client
}

func NewClient(cfg *Config) (*Client, error) {
	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.User, cfg.Password, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	c := &Client{
		config: cfg,
		client: client,
	}

	if err := c.ensureBucketExists(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) ensureBucketExists() error {
	ctx := context.Background()
	exists, err := c.client.BucketExists(ctx, c.config.Bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, c.config.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}

	return nil
}

func (c *Client) Upload(ctx context.Context, key string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.config.Bucket, key, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	return nil
}

func (c *Client) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	reqParams := url.Values{}
	reqParams.Set("response-content-disposition", "attachment")

	url, err := c.client.PresignedGetObject(ctx, c.config.Bucket, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("generate presigned url: %w", err)
	}

	return url.String(), nil
}

func (c *Client) Download(ctx context.Context, key string) (io.Reader, error) {
	obj, err := c.client.GetObject(ctx, c.config.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	return obj, nil
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.StatObject(ctx, c.config.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	err := c.client.RemoveObject(ctx, c.config.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (c *Client) Stat(ctx context.Context, key string) (minio.ObjectInfo, error) {
	info, err := c.client.StatObject(ctx, c.config.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("stat object: %w", err)
	}
	return info, nil
}

func (c *Client) GetClient() *minio.Client {
	return c.client
}

func (c *Client) GetBucket() string {
	return c.config.Bucket
}
