package reconciliation

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
)

// UploadFile uploads reconciliation file to gcs and validate header file
func (r *ReconciliationService) UploadFile(ctx context.Context, transactionType, createdBy string, fileRecon io.Reader, fileHeader *multipart.FileHeader) (*string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/UploadFile")
	defer segment.End()

	rawFile, _ := io.ReadAll(fileRecon)
	defer func() { rawFile = nil }()

	f, err := r.excel.OpenReader(bytes.NewBuffer(rawFile))
	if err != nil {
		r.logger.Error(ctx, "Failed to open reconciliation file reader", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrOpenFileReader)
	}
	defer f.Close()

	_, err = r.getRowsAndValidateBulkUpload(f)
	if err != nil {
		return nil, err
	}

	objectName := filepath.Join(
		constant.ReconciliationUploadDir,
		fmt.Sprintf(
			constant.DefaultFilenameUploadReconciliation,
			util.GetCurrentTimeWithMillisFormatted(),
		),
	) + constant.DefaultExtXlsx

	// Add Content-Disposition header to force download in browsers
	filename := filepath.Base(objectName)
	gcsFilePath, err := r.gcs.UploadFile(ctx, objectName, bytes.NewBuffer(rawFile), true, gcs.WriteContentDisposition(fmt.Sprintf("attachment; filename=\"%s\"", filename)))
	if err != nil {
		r.logger.Error(ctx, "Failed upload reconciliation file to GCS", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	reconData, err := reconciliation.NewReconciliation(transactionType, createdBy, gcsFilePath.ObjectName)
	if err != nil {
		r.logger.Error(ctx, "Failed create reconciliation data", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	if reconData != nil && fileHeader != nil {
		filename := fileHeader.Filename
		if len(filename) >= MAX_FILENAME_LENGTH {
			filename = filename[:MAX_FILENAME_LENGTH]
		}
		reconData.OriginalName = filename
	}

	if err := r.reconRepo.Create(ctx, reconData); err != nil && err != constant.ErrNoRowsAffected {
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	message := map[string]any{
		"uuid": reconData.UUID,
	}

	if err := r.rabbitMqExt.Publish(ctx, rabbitMqExt.ReconProcessRoutingKey, nil, message); err != nil {
		r.logger.Error(ctx, "Failed publish message to rabbitmq", logger.Error(err), logger.String("uuid", reconData.UUID))
	}

	return &reconData.UUID, nil
}
