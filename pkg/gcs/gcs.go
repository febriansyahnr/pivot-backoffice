package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("GCS")

type IGCSService interface {
	Close() error

	SetClient(ctx context.Context) (*storage.Client, error)
	SetBucketWriter(ctx context.Context, objectName string) (*storage.Writer, error)
	UploadFileToGCS(ctx context.Context, objectName, srcFile string, sync bool, ttl *time.Duration) (*Response, error)
	// Deprecated: This function is no longer supported and there is an issue. Please use CreateSignURL as a replacement
	GenerateSignedURL(ctx context.Context, objectName string, ttl *time.Duration) (*Response, error)

	UploadReportingFile(ctx context.Context, filePath, objectName string) (string, error)
	UploadBulkDisbursementFile(ctx context.Context, filePath, objectName string) (*Response, error)
	UploadProofOfTransfer(ctx context.Context, file *multipart.FileHeader, sync bool) (url string, err error)

	// Improvement
	UploadFileFromMultipart(ctx context.Context, objectName string, file *multipart.FileHeader, sync bool) (*UploadMultipart, error)
	UploadFileFromMultipartToBucket(ctx context.Context, bucketName, objectName string, file *multipart.FileHeader, sync bool) (*UploadMultipart, error)
	CreateSignedURL(ctx context.Context, object string, expires time.Duration) (url string, err error)
	ReadAll(ctx context.Context, bucket, object string) ([]byte, error)
	UploadFile(ctx context.Context, objectName string, src io.Reader, sync bool, opts ...WriterOpts) (*UploadMultipart, error)
}

type WriterOpts func(*storage.Writer)

type gcsService struct {
	config Config

	gcsClient *storage.Client
	gcsWriter *storage.Writer
}

const publicURL = "https://storage.googleapis.com"

const (
	PrivateCache = "private"

	ErrStorageNewClientFormat = "storage.NewClient: %v"
	ErrBucketSignedURLFormat  = "Bucket(%q).SignedURL: %w"
	ErrCopyFormat             = "ERROR GCS: io.Copy: %v"
)

// validateFilePath checks if the provided file path is safe and contains no traversal components
// Returns the same path if it's safe or an error if it's potentially unsafe
func validateFilePath(filePath string) (string, error) {
	// Check if the path is empty
	if filePath == "" {
		return "", errors.New("empty file path")
	}

	// Convert to absolute path and normalize
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %v", err)
	}

	// Check for suspicious patterns in the path
	suspicious := regexp.MustCompile(`(\\\.+\\|/\.\./|\.\./|\.\.\\)`)
	if suspicious.MatchString(filePath) {
		return "", errors.New("potential directory traversal detected")
	}

	// Check if the file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file does not exist: %s", absPath)
		}
		return "", fmt.Errorf("error accessing file: %v", err)
	}

	// Ensure it's a regular file, not a directory or something else
	if !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", absPath)
	}

	return absPath, nil
}

func NewGCSService(config Config) IGCSService {
	once.Do(func() {
		client, rootErr = storage.NewClient(context.Background())
	})

	return &gcsService{
		config: config,
	}
}

func (g *gcsService) SetClient(ctx context.Context) (*storage.Client, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/InitClient")
	defer segment.End()

	if g.gcsClient != nil {
		return g.gcsClient, nil
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(ErrStorageNewClientFormat, err)
	}

	// Assign gcsClient
	g.gcsClient = client

	return g.gcsClient, nil
}

func (g *gcsService) SetBucketWriter(ctx context.Context, objectName string) (*storage.Writer, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/InitBucketWriter")
	defer segment.End()

	if g.gcsClient == nil {
		if _, err := g.SetClient(ctx); err != nil {
			return nil, err
		}
	}

	// Create and assign writer
	g.gcsWriter = g.gcsClient.Bucket(g.config.ServiceBucketName).Object(objectName).NewWriter(ctx)
	g.gcsWriter.ContentType = g.mimeTypes(strings.Replace(filepath.Ext(objectName), ".", "", 1))

	return g.gcsWriter, nil
}

// Deprecated: This function is no longer supported and there is an issue. Please use CreateSignURL as a replacement
func (g *gcsService) GenerateSignedURL(ctx context.Context, objectName string, ttl *time.Duration) (*Response, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/GenerateSignedURL")
	defer segment.End()

	if g.gcsClient == nil {
		if _, err := g.SetClient(ctx); err != nil {
			return nil, err
		}
	}
	defer g.gcsClient.Close()

	if g.gcsWriter == nil {
		if _, err := g.SetBucketWriter(ctx, objectName); err != nil {
			return nil, err
		}
	}
	defer g.gcsWriter.Close()

	// Construct the public URL
	publicURL := fmt.Sprintf("%s/%s/%s", publicURL, g.config.ServiceBucketName, objectName)

	expireDuration := 15 * time.Minute
	if ttl != nil {
		expireDuration = *ttl
	}

	signedUrl, err := g.gcsClient.Bucket(g.config.ServiceBucketName).SignedURL(objectName, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expireDuration),
	})
	if err != nil {
		return nil, fmt.Errorf(ErrBucketSignedURLFormat, g.config.ServiceBucketName, err)
	}

	return &Response{
		PublicUrl: publicURL,
		SignedUrl: signedUrl,
	}, nil
}

func (g *gcsService) UploadFileToGCS(ctx context.Context, objectName, srcFile string, sync bool, ttl *time.Duration) (*Response, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/UploadFileToGCS")
	defer segment.End()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(ErrStorageNewClientFormat, err)
	}

	chanErr := make(chan error, 1)
	go func() {
		defer client.Close()

		// Validate file path before opening
		validatedPath, err := validateFilePath(srcFile)
		if err != nil {
			chanErr <- fmt.Errorf("file path validation failed: %v", err)
			fmt.Printf("ERROR: file path validation failed: %v\n", err)
			return
		}

		// #nosec G304 - Path has been validated by validateFilePath function to prevent traversal attacks
		f, err := os.Open(validatedPath)
		if err != nil {
			chanErr <- fmt.Errorf("file.Open: %v", err)
			fmt.Printf("ERROR: file.Open: %v\n", err)
			return
		}
		defer f.Close()

		if !sync {
			ctx = context.Background()
		}
		wr := client.Bucket(g.config.ServiceBucketName).Object(objectName).NewWriter(ctx)
		wr.CacheControl = "private"
		wr.Created = time.Now().UTC()
		wr.ContentType = g.mimeTypes(strings.Replace(filepath.Ext(objectName), ".", "", 1))

		if _, err := io.Copy(wr, f); err != nil {
			chanErr <- fmt.Errorf(ErrCopyFormat, err)
			fmt.Printf("ERROR: io.Copy: %v\n", err)
			return
		}

		chanErr <- wr.Close()
	}()

	if sync {
		if err := <-chanErr; err != nil {
			return nil, err
		}
	}

	expireDuration := 15 * time.Minute
	if ttl != nil {
		expireDuration = *ttl
	}

	signedUrl, err := client.Bucket(g.config.ServiceBucketName).SignedURL(objectName, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expireDuration),
	})
	if err != nil {
		return nil, fmt.Errorf(ErrBucketSignedURLFormat, g.config.ServiceBucketName, err)
	}

	return &Response{
		PublicUrl: publicURL,
		SignedUrl: signedUrl,
	}, nil
}

func (g gcsService) UploadReportingFile(
	ctx context.Context,
	filePath, objectName string,
) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/UploadReportingFile")
	defer segment.End()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf(ErrStorageNewClientFormat, err)
	}
	defer client.Close()

	// Validate file path before opening
	validatedPath, err := validateFilePath(filePath)
	if err != nil {
		return "", fmt.Errorf("file path validation failed: %v", err)
	}

	// #nosec G304 - Path has been validated by validateFilePath function to prevent traversal attacks
	f, err := os.Open(validatedPath)
	if err != nil {
		return "", fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	wc := client.Bucket(g.config.ReportingBucketName).Object(objectName).NewWriter(ctx)
	// Set the content type to Excel
	wc.ContentType = g.mimeTypes(strings.Replace(filepath.Ext(objectName), ".", "", 1)) // For XLSX format

	if _, err = io.Copy(wc, f); err != nil {
		return "", fmt.Errorf(ErrCopyFormat, err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}

	signedUrl, err := client.Bucket(g.config.ReportingBucketName).SignedURL(objectName, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return "", fmt.Errorf(ErrBucketSignedURLFormat, g.config.ReportingBucketName, err)
	}

	return signedUrl, nil
}

func (g gcsService) UploadBulkDisbursementFile(
	ctx context.Context,
	filePath, objectName string,
) (*Response, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/UploadBulkDisbursementFile")
	defer segment.End()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(ErrStorageNewClientFormat, err)
	}
	defer client.Close()

	// Validate file path before opening
	validatedPath, err := validateFilePath(filePath)
	if err != nil {
		return nil, fmt.Errorf("file path validation failed: %v", err)
	}

	// #nosec G304 - Path has been validated by validateFilePath function to prevent traversal attacks
	f, err := os.Open(validatedPath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	wc := client.Bucket(g.config.BulkDisbursementBucketName).Object(objectName).NewWriter(ctx)
	// Set the content type to Excel
	wc.ContentType = g.mimeTypes(strings.Replace(filepath.Ext(objectName), ".", "", 1)) // For XLSX format

	if _, err = io.Copy(wc, f); err != nil {
		return nil, fmt.Errorf(ErrCopyFormat, err)
	}
	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("Writer.Close: %v", err)
	}

	// Construct the public URL
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.config.BulkDisbursementBucketName, objectName)

	signedUrl, err := client.Bucket(g.config.BulkDisbursementBucketName).SignedURL(objectName, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return nil, fmt.Errorf(ErrBucketSignedURLFormat, g.config.BulkDisbursementBucketName, err)
	}

	return &Response{
		PublicUrl: publicURL,
		SignedUrl: signedUrl,
	}, nil
}

func (s *gcsService) UploadProofOfTransfer(ctx context.Context, file *multipart.FileHeader, sync bool) (url string, err error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/UploadProofOfTransfer")
	defer segment.End()

	resp, err := s.UploadFileFromMultipart(
		ctx, s.config.ProofOfTransferFolderName+"/"+file.Filename, file, sync,
	)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", resp.PublicURL, resp.Bucket, resp.ObjectName), nil
}

func (s *gcsService) UploadFileFromMultipart(ctx context.Context, objectName string, file *multipart.FileHeader, sync bool) (*UploadMultipart, error) {
	return s.UploadFileFromMultipartToBucket(ctx, s.config.ServiceBucketName, objectName, file, sync)
}

func (s *gcsService) UploadFileFromMultipartToBucket(ctx context.Context, bucketName, objectName string, file *multipart.FileHeader, sync bool) (*UploadMultipart, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/gcs/UploadFileFromMultipartToBucket")
	defer segment.End()

	if client == nil {
		return nil, rootErr
	}
	chanErr := make(chan error, 1)

	go func() {
		f, err := file.Open()
		if err != nil {
			chanErr <- fmt.Errorf("file.Open: %v", err)
			fmt.Printf("ERROR: file.Open: %v\n", err)
			return
		}
		defer f.Close()

		if !sync {
			ctx = context.Background()
		}
		wr := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
		wr.CacheControl = "private"
		wr.Created = time.Now().UTC()
		wr.ContentType = s.mimeTypes(strings.Replace(filepath.Ext(file.Filename), ".", "", 1))

		if _, err := io.Copy(wr, f); err != nil {
			chanErr <- fmt.Errorf(ErrCopyFormat, err)
			fmt.Printf("ERROR: io.Copy: %v\n", err)
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
		PublicURL:  publicURL,
		Bucket:     bucketName,
		ObjectName: objectName,
	}, nil
}

func (s *gcsService) mimeTypes(ext string) string {
	switch ext {
	case "png":
		return "image/png"

	case "jpeg", "jpg":
		return "image/jpeg"

	case "pdf":
		return "application/pdf"

	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	default:
		return ""
	}
}
