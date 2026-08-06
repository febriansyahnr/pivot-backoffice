package merchant

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// MerchantShortNameValidation validates the provided sub-merchant short name against reserved short names.
// It checks if the short name is empty, and if so, returns an error indicating the short name is invalid.
// If the short name is found in the reserved list (stored in Redis), it checks whether the parent merchant
// is allowed to use the reserved short name. If allowed, it returns nil; otherwise, it returns an error
// indicating the short name is reserved. Any other Redis errors are logged and returned.
func (s *MerchantService) MerchantShortNameValidation(ctx context.Context, shortName string, parentMerchantID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/SubMerchantShortNameValidation")
	defer segment.End()

	if shortName == "" {
		s.logger.Info(ctx, "merchant short name should not empty")
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantShortNameInvalid)
	}

	err := s.CheckOrSetReservedShortNames(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to check key existance", logger.Error(err))
		return err
	}

	var val string
	err = s.redis.HGetScan(ctx, constant.MerchantReservedShortNameKey, strings.ToUpper(shortName), &val)
	if errors.Is(err, redisExt.ErrNil) {
		s.logger.Info(ctx, "short name is not reserved")
		return nil
	}

	if err != nil {
		s.logger.Error(ctx, "failed to get short name rule", logger.Error(err), logger.String("shortname", shortName), logger.String("merchantID", parentMerchantID))
		return err
	}

	allowedMerchants := strings.Split(val, ",")
	for _, merchantID := range allowedMerchants {
		// when the parent merchant was allowed to use the reserved shortname
		// then return nil immediately
		if merchantID == parentMerchantID {
			return nil
		}
	}

	s.logger.Info(ctx, "the short name is reserved", logger.String("shortname", shortName), logger.String("merchantID", parentMerchantID))
	return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantReservedShortName)
}

// CheckOrSetReservedShortNames checks if the reserved short names for merchants are already set in Redis.
// If not, it reads the reserved short names from the source and sets them in Redis.
// Returns an error if any operation fails during the process.
func (s *MerchantService) CheckOrSetReservedShortNames(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/CheckReservedShortNames")
	defer segment.End()

	exist, err := s.redis.Exists(ctx, constant.MerchantReservedShortNameKey).Result()
	if err != nil {
		s.logger.Error(ctx, "failed to delete cached reserved ", logger.Error(err))
		return err
	}

	if exist > 0 {
		return nil
	}

	reservedShortNames, err := s.ReadReservedShortName(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to read reserved short name", logger.Error(err))
		return err
	}

	err = s.SetReservedSubMerchantShortName(ctx, reservedShortNames)
	if err != nil {
		s.logger.Error(ctx, "failed to set reserved short name", logger.Error(err))
		return err
	}

	return nil
}

// SetReservedSubMerchantShortName stores a list of reserved merchant short names and their allowed merchants in Redis.
// If the request slice is empty, it will store an empty value for the reserved short names key.
// Each item in the request contains a short name and a list of allowed merchants, which are joined into a comma-separated string.
// Returns an error if the operation fails to write to Redis.
func (s *MerchantService) SetReservedSubMerchantShortName(ctx context.Context, request []merchant.ReservedMerchantShortNameItem) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/SetReservedSubMerchantShortName")
	defer segment.End()

	if len(request) == 0 {
		s.logger.Info(ctx, "the reserved shortname is empty, will continue to store it")
		err := s.redis.HSet(ctx, constant.MerchantReservedShortNameKey).Err()
		if err != nil {
			s.logger.Error(ctx, "failed to set merchant reserved short name", logger.Error(err))
			return err
		}

		return nil
	}

	values := []any{}
	for _, item := range request {
		values = append(values, item.ShortName, strings.Join(item.AllowedMerchants, ","))
	}

	err := s.redis.HSet(ctx, constant.MerchantReservedShortNameKey, values...).Err()
	if err != nil {
		s.logger.Error(ctx, "failed to set merchant reserved short name", logger.Error(err))
		return err
	}

	return nil
}

// ReadReservedShortName reads a reserved merchant short name list from a file stored in Google Cloud Storage (GCS).
// It opens the file as an Excel document, parses its rows, and constructs a slice of ReservedMerchantShortNameItem.
// Each row is expected to contain a short name and a comma-separated list of allowed merchants.
// Returns the list of reserved short names or an error if the file cannot be read or parsed.
func (s *MerchantService) ReadReservedShortName(ctx context.Context) ([]merchant.ReservedMerchantShortNameItem, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/ReadReservedShortName")
	defer segment.End()
	var (
		reservedShortName = []merchant.ReservedMerchantShortNameItem{}
	)

	objectName := fmt.Sprintf("%s/%s", s.config.GCSConfig.MerchantReservedSortName, constant.MerchantReservedShortNameDefaultFileName)
	fileBytes, err := s.gcs.ReadAll(ctx, s.config.GCSConfig.ServiceBucketName, objectName)
	if err != nil {
		s.logger.Error(ctx, "failed to read file in bucket", logger.Error(err))
		return reservedShortName, err
	}

	excelFile, err := s.excel.OpenReader(bytes.NewBuffer(fileBytes))
	if err != nil {
		s.logger.Error(ctx, "failed to open file reader", logger.Error(err))
		return reservedShortName, err
	}
	defer excelFile.Close()

	rows, err := excelFile.GetRows("data", xlsx.Options{RawCellValue: true})
	if err != nil {
		s.logger.Error(ctx, "failed to get file rows", logger.Error(err))
		return reservedShortName, err
	}

	if len(rows) < 2 {
		s.logger.Info(ctx, "reserved short name file was empty")
		return reservedShortName, nil
	}

	// avoid the header
	for _, row := range rows[1:] {
		allowedMerchants := []string{}
		if len(row) == 3 {
			allowedMerchants = strings.Split(strings.ReplaceAll(row[2], " ", ""), ",")
		}

		reservedShortName = append(reservedShortName, merchant.ReservedMerchantShortNameItem{
			ShortName:        strings.ToUpper(row[0]),
			AllowedMerchants: allowedMerchants,
		})
	}

	return reservedShortName, nil
}

func (s *MerchantService) UploadReservedShortName(ctx context.Context, request *merchant.ReservedMerchantShortNameRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/ReadReservedShortName")
	defer segment.End()

	err := s.redis.Del(ctx, constant.MerchantReservedShortNameKey).Err()
	if err != nil {
		s.logger.Error(ctx, "failed to delete cached reserved ", logger.Error(err))
		return err
	}

	err = s.BackupActiveReservedShortName(ctx, request)
	if err != nil {
		s.logger.Error(ctx, "failed to backup reserved shortname", logger.Error(err))
		return err
	}

	reservedShortNames, err := s.ReadReservedShortName(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to read reserved short name", logger.Error(err))
		return err
	}

	err = s.SetReservedSubMerchantShortName(ctx, reservedShortNames)
	if err != nil {
		s.logger.Error(ctx, "failed to set reserved short name", logger.Error(err))
		return err
	}

	return nil
}

func (s *MerchantService) BackupActiveReservedShortName(ctx context.Context, request *merchant.ReservedMerchantShortNameRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/BackupActiveReservedShortName")
	defer segment.End()

	activeReservedFile := fmt.Sprintf("%s/%s", s.config.GCSConfig.MerchantReservedSortName, constant.MerchantReservedShortNameDefaultFileName)
	fileBytes, err := s.gcs.ReadAll(ctx, s.config.GCSConfig.ServiceBucketName, activeReservedFile)
	if err != nil {
		s.logger.Error(ctx, "failed to read file in bucket", logger.Error(err))
		return err
	}

	excelFile, err := s.excel.OpenReader(bytes.NewBuffer(fileBytes))
	if err != nil {
		s.logger.Error(ctx, "failed to open file reader", logger.Error(err))
		return err
	}
	defer excelFile.Close()

	// override active reserved short name
	_, err = s.gcs.UploadFileFromMultipart(ctx, activeReservedFile, request.File, true)
	if err != nil {
		s.logger.Error(ctx, "failed to upload multipart form", logger.Error(err))
		return err
	}

	buff, err := excelFile.WriteToBuffer()
	if err != nil {
		s.logger.Error(ctx, "failed to write buffer from xlsx file", logger.Error(err))
		return err
	}

	backupName := fmt.Sprintf(constant.MerchantReservedShortNameBackupFileName, time.Now().Format(time.RFC3339))
	backupObjectName := fmt.Sprintf("%s/%s", s.config.GCSConfig.MerchantReservedSortName, backupName)
	writer, err := s.gcs.SetBucketWriter(ctx, backupObjectName)
	if err != nil {
		s.logger.Error(ctx, "failed to create gcs writer", logger.Error(err))
		return err
	}

	_, err = writer.Write(buff.Bytes())
	if err != nil {
		writer.Close()
		s.logger.Error(ctx, "failed to write buffer into gcs bucket writer")
		return err
	}

	writer.ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	err = writer.Close()
	if err != nil {
		s.logger.Error(ctx, "failed to close gcs bucket writer")
		return err
	}

	return nil
}
