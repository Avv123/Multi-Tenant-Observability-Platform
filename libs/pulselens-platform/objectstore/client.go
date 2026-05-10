package objectstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Client struct {
	enabled bool
	bucket  string
	prefix  string
	client  *s3.Client
}

func New(enabled bool, endpoint, region, accessKey, secretKey, bucket, prefix string, forcePathStyle bool) (*Client, error) {
	bucket = strings.TrimSpace(bucket)
	if !enabled || bucket == "" {
		return &Client{enabled: false, bucket: bucket, prefix: cleanPrefix(prefix)}, nil
	}

	if strings.TrimSpace(region) == "" {
		region = "us-east-1"
	}
	if strings.TrimSpace(accessKey) == "" {
		accessKey = "test"
	}
	if strings.TrimSpace(secretKey) == "" {
		secretKey = "test"
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}
	if strings.TrimSpace(endpoint) != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(func(service string, region string, _ ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})
		loadOptions = append(loadOptions, awsconfig.WithEndpointResolverWithOptions(resolver))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = forcePathStyle
	})

	return &Client{
		enabled: true,
		bucket:  bucket,
		prefix:  cleanPrefix(prefix),
		client:  client,
	}, nil
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled && c.client != nil && c.bucket != ""
}

func (c *Client) Bucket() string {
	if c == nil {
		return ""
	}
	return c.bucket
}

func (c *Client) EnsureBucket(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}

	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(c.bucket)})
	if err == nil {
		return nil
	}

	var notFound *types.NotFound
	if errors.As(err, &notFound) || strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(strings.ToLower(err.Error()), "404") {
		_, createErr := c.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(c.bucket)})
		return createErr
	}
	if strings.Contains(strings.ToLower(err.Error()), "status code: 301") {
		return nil
	}
	return err
}

func (c *Client) PutObject(ctx context.Context, key string, payload []byte, contentType string) (string, error) {
	if !c.Enabled() {
		return "", nil
	}

	objectKey := c.objectKey(key)
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(payload),
		ContentType: aws.String(strings.TrimSpace(contentType)),
	})
	if err != nil {
		return "", err
	}
	return objectKey, nil
}

func (c *Client) GetObject(ctx context.Context, bucket string, key string) ([]byte, error) {
	if !c.Enabled() {
		return nil, nil
	}

	targetBucket := strings.TrimSpace(bucket)
	if targetBucket == "" {
		targetBucket = c.bucket
	}
	response, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(targetBucket),
		Key:    aws.String(strings.TrimSpace(key)),
	})
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return io.ReadAll(response.Body)
}

func (c *Client) DeleteObject(ctx context.Context, bucket string, key string) error {
	if !c.Enabled() {
		return nil
	}

	targetBucket := strings.TrimSpace(bucket)
	if targetBucket == "" {
		targetBucket = c.bucket
	}
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(targetBucket),
		Key:    aws.String(strings.TrimSpace(key)),
	})
	return err
}

func (c *Client) URI(key string) string {
	if c == nil || c.bucket == "" || strings.TrimSpace(key) == "" {
		return ""
	}
	return fmt.Sprintf("s3://%s/%s", c.bucket, strings.TrimSpace(key))
}

func (c *Client) objectKey(key string) string {
	key = strings.Trim(strings.TrimSpace(key), "/")
	if c.prefix == "" {
		return key
	}
	if key == "" {
		return c.prefix
	}
	return c.prefix + "/" + key
}

func cleanPrefix(prefix string) string {
	return strings.Trim(strings.TrimSpace(prefix), "/")
}
