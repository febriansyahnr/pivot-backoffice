package merchant

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const (
	defaultMerchantLogoBucket = "stg-merchant-logo" // default bucket if not configured
)

// UploadMerchantLogo uploads a merchant logo to GCS and updates the merchant logo URL in the database
func (s *MerchantService) UploadMerchantLogo(ctx context.Context, merchantID string, file *multipart.FileHeader) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UploadMerchantLogo")
	defer segment.End()

	// Validate merchant exists
	merchant, err := s.repo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "error when find merchant by id", logger.Error(err))
		return "", pkgErrors.New(responseHttp.HttpErrInternal, err)
	}
	if merchant == nil {
		return "", pkgErrors.New(responseHttp.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	// Validate file
	if err := s.validateLogoFile(file); err != nil {
		return "", pkgErrors.New(responseHttp.HttpErrRequest, err)
	}

	// Get bucket name from config, use default if not set
	bucketName := defaultMerchantLogoBucket
	if s.config != nil && s.config.GCSConfig.MerchantLogoBucketName != "" {
		bucketName = s.config.GCSConfig.MerchantLogoBucketName
	}

	// Generate unique filename (without folder prefix since we're using a dedicated bucket)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	objectName := fmt.Sprintf("%s_%d%s", merchantID, time.Now().Unix(), ext)

	// Upload to GCS with public ACL
	uploadResp, err := s.uploadPublicFile(ctx, bucketName, objectName, file)
	if err != nil {
		s.logger.Error(ctx, "error when upload merchant logo to GCS", logger.Error(err))
		return "", pkgErrors.New(responseHttp.HttpErrInternal, fmt.Errorf("failed to upload logo"))
	}

	// Construct the full public URL
	logoURL := fmt.Sprintf("%s/%s/%s", uploadResp.PublicURL, uploadResp.Bucket, uploadResp.ObjectName)

	// Update merchant logo in database
	merchant.Logo = logoURL
	merchant.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, merchant); err != nil {
		s.logger.Error(ctx, "error when update merchant logo", logger.Error(err))
		return "", pkgErrors.New(responseHttp.HttpErrInternal, fmt.Errorf("failed to update merchant logo"))
	}

	// Clear merchant status cache
	cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchant.UUID)
	_ = s.redis.Del(ctx, cacheKey)

	return logoURL, nil
}

// uploadPublicFile uploads a file to the public GCS bucket
func (s *MerchantService) uploadPublicFile(ctx context.Context, bucketName, objectName string, file *multipart.FileHeader) (*gcsResponse, error) {
	// Upload to GCS (bucket is already public, no need to set object ACL)
	uploadResp, err := s.gcs.UploadFileFromMultipartToBucket(ctx, bucketName, objectName, file, true)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &gcsResponse{
		PublicURL:  uploadResp.PublicURL,
		Bucket:     uploadResp.Bucket,
		ObjectName: uploadResp.ObjectName,
	}, nil
}

type gcsResponse struct {
	PublicURL  string
	Bucket     string
	ObjectName string
}

// validateLogoFile validates the uploaded logo file
func (s *MerchantService) validateLogoFile(file *multipart.FileHeader) error {
	if file == nil {
		return fmt.Errorf("file is required")
	}

	// Check file size
	if file.Size > constant.FileSize5MB {
		return fmt.Errorf("file size exceeds maximum limit of 5MB")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("invalid file type. Only PNG, JPG, and JPEG are allowed")
	}

	// Validate filename has no suspicious characters
	filename := filepath.Base(file.Filename)
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename")
	}

	return nil
}
