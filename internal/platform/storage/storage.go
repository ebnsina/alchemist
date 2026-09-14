// Package storage wraps S3-compatible object storage. MinIO in dev, SeaweedFS in
// production, same interface.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type Store struct {
	client   *s3.Client
	presign  *s3.PresignClient
	uploader *manager.Uploader
	bucket   string
}

func New(ctx context.Context, endpoint, region, bucket, accessKey, secretKey string) (*Store, error) {
	cfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	// Path-style addressing is required for MinIO and SeaweedFS.
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	// Rendition files run to hundreds of megabytes for long content. A single
	// PutObject caps at 5GB and grows unreliable well before that, so uploads go
	// through the transfer manager, which splits into multipart parts and retries
	// each part independently.
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 16 * 1024 * 1024
		u.Concurrency = 4
	})

	return &Store{
		client:   client,
		presign:  s3.NewPresignClient(client),
		uploader: uploader,
		bucket:   bucket,
	}, nil
}

func (s *Store) Put(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put %s: %w", key, err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", key, err)
	}
	return out.Body, nil
}

// GetRange fetches a byte range. Chunk workers use this to read only the slice of
// the mezzanine they need, which is why no shared filesystem is required.
func (s *Store) GetRange(ctx context.Context, key string, start, end int64) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", start, end)),
	})
	if err != nil {
		return nil, fmt.Errorf("get range %s [%d-%d]: %w", key, start, end, err)
	}
	return out.Body, nil
}

// PresignPut returns a URL the client uploads to directly, so source bytes never
// pass through the API.
func (s *Store) PresignPut(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign put %s: %w", key, err)
	}
	return req.URL, nil
}

// Object carries the parts of an S3 response the origin has to pass through
// verbatim for range requests and conditional requests to work.
type Object struct {
	Body          io.ReadCloser
	ETag          string
	ContentLength int64
	ContentRange  string
}

var (
	ErrNotFound            = errors.New("object not found")
	ErrRangeNotSatisfiable = errors.New("range not satisfiable")
)

// GetPassthrough forwards a client Range header to storage and returns the response
// metadata unchanged, so the origin never has to buffer a whole object to serve a
// slice of it.
func (s *Store) GetPassthrough(ctx context.Context, key, rangeHeader string) (*Object, error) {
	in := &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}
	if rangeHeader != "" {
		in.Range = aws.String(rangeHeader)
	}
	out, err := s.client.GetObject(ctx, in)
	if err != nil {
		var nsk *types.NoSuchKey
		var nf *types.NotFound
		if errors.As(err, &nsk) || errors.As(err, &nf) {
			return nil, ErrNotFound
		}
		// A player seeking past the end must get 416, not a 5xx: nginx caches and
		// retries 5xx, and a range error is the client's to correct, not ours.
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidRange" {
			return nil, ErrRangeNotSatisfiable
		}
		return nil, fmt.Errorf("get %s: %w", key, err)
	}
	o := &Object{Body: out.Body}
	if out.ETag != nil {
		o.ETag = *out.ETag
	}
	if out.ContentLength != nil {
		o.ContentLength = *out.ContentLength
	}
	if out.ContentRange != nil {
		o.ContentRange = *out.ContentRange
	}
	return o, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}

// Listing is one page of objects from a customer's bucket.
type Listing struct {
	Objects []ListedObject
	Cursor  string
	HasMore bool
}

type ListedObject struct {
	Key       string
	ETag      string
	Size      int64
	Multipart bool
}

// NewClient builds a store against someone else's credentials, for reading a
// customer-owned bucket.
func NewClient(ctx context.Context, endpoint, region, bucket, accessKey, secretKey string) (*Store, error) {
	return New(ctx, endpoint, region, bucket, accessKey, secretKey)
}

// List returns one page of objects under prefix, resuming from cursor.
//
// Paged rather than exhaustive on purpose: a customer library can hold millions of
// objects, and a reconcile that must finish before it makes progress never finishes.
func (s *Store) List(ctx context.Context, prefix, cursor string, limit int32) (*Listing, error) {
	in := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		MaxKeys: aws.Int32(limit),
	}
	if prefix != "" {
		in.Prefix = aws.String(prefix)
	}
	if cursor != "" {
		in.ContinuationToken = aws.String(cursor)
	}

	out, err := s.client.ListObjectsV2(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", s.bucket, err)
	}

	l := &Listing{}
	for _, o := range out.Contents {
		if o.Key == nil {
			continue
		}
		obj := ListedObject{Key: *o.Key}
		if o.ETag != nil {
			obj.ETag = strings.Trim(*o.ETag, `"`)
		}
		if o.Size != nil {
			obj.Size = *o.Size
		}
		// A multipart object's ETag ends in "-<partcount>". Harmless to ingest, but
		// worth flagging so a still-uploading object can be skipped.
		obj.Multipart = strings.Contains(obj.ETag, "-")
		l.Objects = append(l.Objects, obj)
	}
	if out.NextContinuationToken != nil {
		l.Cursor = *out.NextContinuationToken
	}
	l.HasMore = out.IsTruncated != nil && *out.IsTruncated
	return l, nil
}
