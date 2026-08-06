package amlservice

import (
	"context"
	"encoding/json"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *AmlService) UpdateDetailStatusByProfileId(ctx context.Context, profileID string, merchantID string, request *amlcommon.UpdateDetailStatusRequest) error {
	ctx, span := tracer.Start(ctx, "internal/service/v1/aml/UpdateDetailStatusByProfileId")
	defer span.End()

	if request.Name == "" || request.DOB == "" || request.Status == "" {
		s.logger.Error(ctx, "name, dob, and status are required")
		return constant.ErrInvalidRequestPayload
	}

	if request.Status != amlcommon.DetailStatusPending && request.Status != amlcommon.DetailStatusDismiss && request.Status != amlcommon.DetailStatusOnMonitor {
		s.logger.Error(ctx, "invalid status", logger.String("status", request.Status))
		return constant.ErrInvalidRequestPayload
	}

	merchant, err := s.merchantRepository.FindMerchantByID(ctx, merchantID)
	if err != nil || merchant == nil {
		s.logger.Error(ctx, "merchant not found", logger.Error(err), logger.String("merchantID", merchantID))
		return constant.ErrMerchantNotFound
	}

	screeningKey := request.Name + ":" + request.DOB

	var screeningData commonModel.ThirdPartyScreeningData
	if !(merchant.ThirdPartyScreeningData.Valid && len(merchant.ThirdPartyScreeningData.JSONText) > 0) {
		s.logger.Error(ctx, "no screening data found for merchant")
		return constant.ErrDataNotFound
	}

	err = json.Unmarshal(merchant.ThirdPartyScreeningData.JSONText, &screeningData)
	if err != nil {
		s.logger.Error(ctx, "failed to parse existing screening data", logger.Error(err))
		return err
	}

	if screeningData.AML == nil {
		s.logger.Error(ctx, "no AML screening data found")
		return constant.ErrDataNotFound
	}

	screeningResponse, exists := screeningData.AML[screeningKey]
	if !exists {
		s.logger.Error(ctx, "no screening data found for person", logger.String("screeningKey", screeningKey))
		return constant.ErrDataNotFound
	}

	if screeningResponse.Result == nil {
		s.logger.Error(ctx, "no screening result found for person", logger.String("screeningKey", screeningKey))
		return constant.ErrDataNotFound
	}

	targetDetail := ""
	detailFound := false
	for i, detail := range screeningResponse.Result.Detail {
		if detail.ProfileID == profileID {
			screeningResponse.Result.Detail[i].Status = request.Status
			targetDetail = detail.Name
			detailFound = true
			break
		}
	}

	if !detailFound {
		s.logger.Error(ctx, "no detail found with matching profileID", logger.String("profileID", profileID))
		return constant.ErrDataNotFound
	}

	screeningData.AML[screeningKey] = screeningResponse

	updatedData, err := json.Marshal(screeningData)
	if err != nil {
		s.logger.Error(ctx, "failed to marshal updated screening data", logger.Error(err))
		return err
	}

	screeningDataJSON := types.NullJSONText{
		JSONText: updatedData,
		Valid:    true,
	}

	err = s.merchantRepository.UpdateThirdPartyScreeningData(ctx, merchantID, screeningDataJSON)
	if err != nil {
		s.logger.Error(ctx, "failed to update merchant screening data", logger.Error(err))
		return err
	}

	s.logger.Info(ctx, "successfully updated detail status",
		logger.String("merchantID", merchantID),
		logger.String("profileID", profileID),
		logger.String("screeningKey", screeningKey),
		logger.String("targetDetail", targetDetail),
		logger.String("status", request.Status))

	return nil
}
