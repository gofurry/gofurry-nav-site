package assets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	env "github.com/gofurry/gofurry-admin/config"
	cos "github.com/tencentyun/cos-go-sdk-v5"
)

type COSStore struct{ client *cos.Client }
type R2Store struct {
	client *s3.Client
	bucket string
}

func New(cfg env.AssetStorageConfig) (*Service, error) {
	s := &Service{}
	if cfg.Primary.Configured() {
		if cfg.Primary.Provider != "cos" {
			return nil, errors.New("primary provider must be cos")
		}
		base, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Primary.Bucket, cfg.Primary.Region))
		if err != nil {
			return nil, errors.New("invalid COS configuration")
		}
		client := cos.NewClient(&cos.BaseURL{BucketURL: base}, &http.Client{Timeout: 8 * time.Second, Transport: &cos.AuthorizationTransport{SecretID: cfg.Primary.AccessKeyID, SecretKey: cfg.Primary.SecretAccessKey}})
		s.Primary = &COSStore{client: client}
	}
	if cfg.Mirror.Configured() {
		endpoint, err := url.Parse(cfg.Mirror.Endpoint)
		if cfg.Mirror.Provider != "r2" || err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.User != nil {
			return nil, errors.New("invalid R2 configuration")
		}
		client := s3.NewFromConfig(aws.Config{Region: cfg.Mirror.Region, Credentials: credentials.NewStaticCredentialsProvider(cfg.Mirror.AccessKeyID, cfg.Mirror.SecretAccessKey, ""), HTTPClient: &http.Client{Timeout: 8 * time.Second}, RetryMaxAttempts: 2}, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Mirror.Endpoint)
			o.UsePathStyle = true
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		})
		s.Mirror = &R2Store{client: client, bucket: cfg.Mirror.Bucket}
	}
	return s, nil
}
func (s *COSStore) Put(ctx context.Context, o Object) error {
	headers := http.Header{}
	headers.Set("x-cos-meta-sha256", o.SHA256)
	headers.Set("x-cos-meta-asset-kind", o.Kind)
	r, err := s.client.Object.Put(ctx, o.Key, bytes.NewReader(o.Data), &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: o.ContentType, CacheControl: CacheControl, XCosMetaXXX: &headers}})
	if err != nil {
		return cosError("PUT", r)
	}
	return nil
}
func (s *COSStore) Head(ctx context.Context, key string) (Info, error) {
	r, err := s.client.Object.Head(ctx, key, nil)
	if r != nil && r.StatusCode == 404 {
		return Info{}, nil
	}
	if err != nil {
		return Info{}, cosError("HEAD", r)
	}
	return Info{Exists: true, Size: r.ContentLength, ContentType: r.Header.Get("Content-Type"), SHA256: r.Header.Get("x-cos-meta-sha256"), Kind: r.Header.Get("x-cos-meta-asset-kind"), CacheControl: r.Header.Get("Cache-Control")}, nil
}
func (s *COSStore) Get(ctx context.Context, key string) ([]byte, error) {
	r, err := s.client.Object.Get(ctx, key, nil)
	if err != nil {
		return nil, cosError("GET", r)
	}
	defer r.Body.Close()
	return readBounded(r.Body)
}
func cosError(operation string, r *cos.Response) error {
	code := 0
	if r != nil {
		code = r.StatusCode
	}
	return fmt.Errorf("COS %s failed (HTTP %d)", operation, code)
}

func (s *R2Store) Put(ctx context.Context, o Object) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: &s.bucket, Key: &o.Key, Body: bytes.NewReader(o.Data), ContentLength: aws.Int64(int64(len(o.Data))), ContentType: &o.ContentType, CacheControl: aws.String(CacheControl), Metadata: map[string]string{"sha256": o.SHA256, "asset-kind": o.Kind}})
	return r2Error(err)
}
func (s *R2Store) Head(ctx context.Context, key string) (Info, error) {
	r, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &s.bucket, Key: &key})
	var responseErr *smithyhttp.ResponseError
	if errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == 404 {
		return Info{}, nil
	}
	if err != nil {
		return Info{}, r2Error(err)
	}
	return Info{Exists: true, Size: aws.ToInt64(r.ContentLength), ContentType: aws.ToString(r.ContentType), SHA256: r.Metadata["sha256"], Kind: r.Metadata["asset-kind"], CacheControl: aws.ToString(r.CacheControl)}, nil
}
func (s *R2Store) Get(ctx context.Context, key string) ([]byte, error) {
	r, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: &s.bucket, Key: &key})
	if err != nil {
		return nil, r2Error(err)
	}
	defer r.Body.Close()
	return readBounded(r.Body)
}
func r2Error(err error) error {
	if err == nil {
		return nil
	}
	var api smithy.APIError
	if errors.As(err, &api) {
		return fmt.Errorf("R2 operation failed (%s)", api.ErrorCode())
	}
	return errors.New("R2 request failed")
}
func readBounded(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxSize+1))
	if err != nil {
		return nil, errors.New("object read failed")
	}
	if len(data) > MaxSize {
		return nil, errors.New("object exceeds size limit")
	}
	return data, nil
}
func URL(base, key string) string {
	if strings.TrimSpace(base) == "" || !ValidKey(key) {
		return ""
	}
	return strings.TrimRight(base, "/") + "/" + key
}
