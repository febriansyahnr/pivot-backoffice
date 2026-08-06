package qris

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	"errors"

	"github.com/google/uuid"
)

func (s *qrisService) DuplicateRegistration(ctx context.Context, request *qris.DuplicateRegistrationReq) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/DuplicateRegistration")
	defer segment.End()

	s.logger.Info(ctx, "Starting duplicate QRIS registration",
		logger.String("sourceMerchantId", request.SourceMerchantId),
		logger.String("targetMerchantId", request.TargetMerchantId))

	// Get source merchant external ID (the one with existing registration)
	sourceMerchant, err := s.merchantRepo.FindMerchantByID(ctx, request.SourceMerchantId)
	if err != nil {
		s.logger.Error(ctx, "Failed to get source merchant by ID", logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}
	if sourceMerchant == nil {
		s.logger.Error(ctx, "Source merchant not found", logger.Any("sourceMerchantId", request.SourceMerchantId))
		return "", pkgErrors.New(httpResponse.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	// Get target merchant external ID (where registration will be duplicated to)
	targetMerchant, err := s.merchantRepo.FindMerchantByID(ctx, request.TargetMerchantId)
	if err != nil {
		s.logger.Error(ctx, "Failed to get target merchant by ID", logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}
	if targetMerchant == nil {
		s.logger.Error(ctx, "Target merchant not found", logger.Any("targetMerchantId", request.TargetMerchantId))
		return "", pkgErrors.New(httpResponse.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	// Check if target merchant already has a QR registration
	targetRegistration, err := s.repository.FindQrRegistrationByExternalID(ctx, targetMerchant.ExternalId)
	if err != nil {
		s.logger.Error(ctx, "Failed to check existing QRIS registration for target merchant", logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}
	if targetRegistration != nil {
		s.logger.Error(ctx, "Target merchant already has a QRIS registration",
			logger.String("targetMerchantId", request.TargetMerchantId),
			logger.String("existingRegistrationId", targetRegistration.Id))
		return "", pkgErrors.New(httpResponse.HttpErrRequest, errors.New("target merchant already has a QRIS registration"))
	}

	// Get QRIS registration from source merchant
	sourceRegistration, err := s.repository.FindQrRegistrationByExternalID(ctx, sourceMerchant.ExternalId)
	if err != nil {
		s.logger.Error(ctx, "Failed to get QRIS registration", logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}
	if sourceRegistration == nil {
		s.logger.Error(ctx, "QRIS registration not found for source merchant",
			logger.String("sourceMerchantId", request.SourceMerchantId),
			logger.String("externalID", sourceMerchant.ExternalId))
		return "", pkgErrors.New(httpResponse.HttpErrNotFound, constant.ErrDataNotFound)
	}

	// Verify the source registration has SUCCESS status
	if sourceRegistration.Status != constant.QrRegistrationStatusSuccess {
		s.logger.Error(ctx, "Source QRIS registration does not have SUCCESS status",
			logger.String("status", sourceRegistration.Status))
		return "", pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrInvalidStatus)
	}

	// Create a new registration with data from the source registration
	newRegistration := &qris.Registration{
		Id:                       uuid.New().String(),
		ExternalId:               targetMerchant.ExternalId,
		Acquirer:                 sourceRegistration.Acquirer,
		MerchantType:             sourceRegistration.MerchantType,
		AcquirerParentMerchantId: sourceRegistration.AcquirerParentMerchantId,
		MerchantName:             targetMerchant.Name, // Use the target merchant's name
		MerchantShortName:        sourceRegistration.MerchantShortName,
		AddressRaw:               sourceRegistration.AddressRaw,
		BusinessInfoRaw:          sourceRegistration.BusinessInfoRaw,
		BusinessDocumentRaw:      sourceRegistration.BusinessDocumentRaw,
		Status:                   constant.QrRegistrationStatusSuccess, // Initialize with SUCCESS status since we're duplicating a successful registration
		AcquirerMerchantId:       sourceRegistration.AcquirerMerchantId,
		CallbackDetailRaw:        sourceRegistration.CallbackDetailRaw,
		CallbackDatetime:         sourceRegistration.CallbackDatetime,
		CreatedAt:                time.Now(),
		CreatedBy:                "system-duplicate",
		UpdatedAt:                time.Now(),
	}

	// Save the new registration
	if err := s.repository.InitRegistration(ctx, newRegistration); err != nil {
		s.logger.Error(ctx, "Failed to save duplicated QRIS registration", logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}

	s.logger.Info(ctx, "Successfully duplicated QRIS registration",
		logger.String("newRegistrationId", newRegistration.Id),
		logger.String("sourceRegistrationId", sourceRegistration.Id))

	return newRegistration.Id, nil
}
