package userLoggedInDeviceService

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userLoggedInDeviceModel "github.com/paper-indonesia/pivot-backoffice/internal/model/userLoggedInDevice"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *UserLoggedInDeviceService) Validate(ctx context.Context, userID, deviceIdentifier string, isRemember bool) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/userLoggedInDevice/Validate")
	defer segment.End()

	// Must have device identifier
	if deviceIdentifier == "" {
		return constant.ErrNeed2FAChallengeForLogin
	}

	existedUserDevice, err := s.userLoggedInDeviceRepo.FindByUserAndDevice(ctx, userID, deviceIdentifier)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// If device still in remember, return nil so no need to do 2fa challenge
	if existedUserDevice != nil {
		if metadata := buildMetadata(existedUserDevice.AdditionalInfo); metadata != nil {
			if metadata.IsRemember && metadata.RememberUntil.After(time.Now().UTC()) {
				return nil
			}
		}
	}

	if existedUserDevice == nil {
		existedUserDevice = &userLoggedInDeviceModel.UserLoggedInDevice{
			UUID:             uuid.NewString(),
			UserID:           userID,
			DeviceIdentifier: deviceIdentifier,
			Status:           constant.UserLoggedInDeviceStatusActive,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}

		if err = s.userLoggedInDeviceRepo.Create(ctx, existedUserDevice); err != nil {
			return pkgErrors.New(response.HttpErrDatabase, err)
		}
	}

	return constant.ErrNeed2FAChallengeForLogin
}

func buildMetadata(additionalInfo *string) *userLoggedInDeviceModel.UserLoggedInDeviceMetadata {
	if additionalInfo == nil {
		return nil
	}

	metadata := userLoggedInDeviceModel.UserLoggedInDeviceMetadata{}

	// Unmarshal the JSON into the struct
	err := json.Unmarshal([]byte(*additionalInfo), &metadata)
	if err != nil {
		return nil
	}

	return &metadata
}
