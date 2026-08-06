package qris

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *qrisService) FindQrRegistrationByExternalID(ctx context.Context, externalID string) (*qris.Registration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/FindQrRegistrationByExternalID")
	defer segment.End()

	qrRegistration, err := s.repository.FindQrRegistrationByExternalID(ctx, externalID)
	if err != nil {
		s.logger.Error(ctx, "Error find QR registration by externalId", logger.Any("externalId", externalID), logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)
	} else if qrRegistration == nil {
		s.logger.Info(ctx, "Find QR registration by externalId is not found", logger.Any("externalId", externalID))
		return nil, pkgErrors.New(httpResponse.HttpErrNotFound, constant.ErrDataNotFound)
	}

	return qrRegistration, nil
}

func (s *qrisService) FindQrRegistrationByExternalIDAndAcquirer(ctx context.Context, externalID string, acquirer string) (*qris.Registration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/FindQrRegistrationByExternalIDAndAcquirer")
	defer segment.End()

	qrRegistration, err := s.repository.FindQrRegistrationByExternalIDAndAcquirer(ctx, externalID, acquirer)
	if err != nil {
		s.logger.Error(ctx, "Error find QR registration by externalId and acquirer",
			logger.Any("externalId", externalID),
			logger.Any("acquirer", acquirer),
			logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}

	if qrRegistration == nil {
		s.logger.Info(ctx, "Find QR registration by externalId and acquirer is not found",
			logger.Any("externalId", externalID),
			logger.Any("acquirer", acquirer))
		return nil, pkgErrors.New(httpResponse.HttpErrNotFound, constant.ErrDataNotFound)
	}

	if qrRegistration.Status != constant.QrRegistrationStatusSuccess {
		s.logger.Warn(ctx, "QR registration status is not success",
			logger.Any("externalId", externalID),
			logger.Any("acquirer", acquirer),
			logger.String("status", qrRegistration.Status))
		return nil, pkgErrors.New(httpResponse.HttpErrUnprocessableContent, constant.ErrQrRegistrationIsNotCompleted)
	}

	return qrRegistration, nil
}
