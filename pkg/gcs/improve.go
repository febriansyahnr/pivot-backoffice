package gcs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
)

var (
	once    sync.Once
	client  *storage.Client
	rootErr error
)

func (gcsService) Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}

func WriteContentDisposition(value string) WriterOpts {
	return func(w *storage.Writer) {
		w.ContentDisposition = value
	}
}

func (s *gcsService) UploadFile(ctx context.Context, objectName string, src io.Reader, sync bool, opts ...WriterOpts) (*UploadMultipart, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/UploadFile")
	defer segment.End()

	if client == nil {
		return nil, rootErr
	}
	chanErr := make(chan error, 1)

	go func() {
		if !sync {
			ctx = context.Background()
		}
		wr := client.Bucket(s.config.ServiceBucketName).Object(objectName).NewWriter(ctx)
		wr.Created = time.Now().UTC()
		wr.CacheControl = PrivateCache
		wr.ContentType = s.mimeTypes(strings.Replace(filepath.Ext(objectName), ".", "", 1))

		for _, opt := range opts {
			opt(wr)
		}

		if _, err := io.Copy(wr, src); err != nil {
			fmt.Printf(ErrCopyFormat+"\n", err)
			chanErr <- fmt.Errorf(ErrCopyFormat, err)
			return
		}

		chanErr <- wr.Close()
	}()

	if sync {
		if err := <-chanErr; err != nil {
			return nil, err
		}
	}
	return &UploadMultipart{
		PublicURL:  path.Join(publicURL, s.config.ServiceBucketName, objectName),
		Bucket:     s.config.ServiceBucketName,
		ObjectName: objectName,
	}, nil
}

func (g *gcsService) CreateSignedURL(ctx context.Context, object string, expires time.Duration) (url string, err error) {
	_, segment := otelTracer.Start(ctx, "pkg/gcs/CreateSignedURL")
	defer segment.End()

	if client == nil {
		return "", rootErr

	} else if expires == 0 {
		expires = 15 * time.Minute
	}

	return client.Bucket(g.config.ServiceBucketName).SignedURL(object, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  http.MethodGet,
		Expires: time.Now().UTC().Add(expires),
	})
}

func (g *gcsService) ReadAll(ctx context.Context, bucket, object string) ([]byte, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/ReadAll")
	defer segment.End()

	if client == nil {
		return nil, rootErr
	}

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("create reader: %w", err)
	}
	defer rc.Close()

	return io.ReadAll(rc)
}
