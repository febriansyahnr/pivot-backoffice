package amlservice

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *AmlService) Profile(ctx context.Context, request *amlcommon.CheckRequest, provider string, merchantID string, profileID string) (*amlcommon.ProfileDetailResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/aml/Check")
	defer span.End()

	referenceID, _ := uuid.NewV7()

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		ReferenceId: referenceID.String(),
		RequestId:   traceID,
		From:        constant.AML_PROCESSOR,
	})

	repo, ok := s.thirdPartyProcessor[provider]
	if !ok {
		err := constant.ErrProviderNotFound
		s.logger.Error(ctx, "provider not found", logger.Error(err), logger.String("provider", provider))
		return nil, err
	}

	// Find screening data to extract inquiryID
	inquiryID, foundProfileID, err := s.findScreeningData(ctx, request, provider, merchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to find screening data", logger.Error(err))
		return nil, err
	}

	if inquiryID == "" {
		err := constant.ErrDataNotFound
		s.logger.Error(ctx, "inquiryID not found in screening data")
		return nil, err
	}

	// If profileID was provided as query parameter, use it; otherwise use the found one
	targetProfileID := foundProfileID
	if profileID != "" {
		targetProfileID = profileID
	}

	if targetProfileID == "" {
		err := constant.ErrDataNotFound
		s.logger.Error(ctx, "profileID not found in screening data and not provided as parameter")
		return nil, err
	}

	// Get profile detail using inquiryID and profileID
	profileResponse, err := repo.ProfileDetail(ctx, inquiryID, targetProfileID)
	if err != nil {
		s.logger.Error(ctx, "failed to get profile detail", logger.Error(err))
		return nil, err
	}

	profileResponse.TransactionID = inquiryID
	profileResponse.ReferenceID = referenceID.String()
	return profileResponse, nil
}

// find or do screening data
func (s *AmlService) findScreeningData(ctx context.Context, request *amlcommon.CheckRequest, provider string, merchantID string) (string, string, error) {
	var screeningKey string
	if request.SubjectType == constant.AML_SUBJECT_TYPE_ENTITY {
		screeningKey = request.Name
	} else {
		screeningKey = request.Name + ":" + request.DOB
	}

	// merchant ID is present, then we can check if theres already AML screening there
	if merchantID != "" {
		merchant, err := s.merchantRepository.FindMerchantByID(ctx, merchantID)
		if err != nil || merchant == nil {
			s.logger.Error(ctx, "merchant not found", logger.Error(err), logger.String("merchantID", merchantID))
			return "", "", constant.ErrMerchantNotFound
		}

		var screeningData commonModel.ThirdPartyScreeningData
		if merchant.ThirdPartyScreeningData.Valid && len(merchant.ThirdPartyScreeningData.JSONText) > 0 {
			err = json.Unmarshal(merchant.ThirdPartyScreeningData.JSONText, &screeningData)
			if err != nil {
				s.logger.Warn(ctx, "failed to parse existing screening data, creating new structure", logger.Error(err))
				screeningData = commonModel.ThirdPartyScreeningData{}
			}
		} else {
			screeningData = commonModel.ThirdPartyScreeningData{}
		}

		if screeningData.AML == nil {
			screeningData.AML = make(map[string]*amlcommon.ScreeningResponse)
		}

		// return the inquiry ID and profileID if exists
		if existingData, exists := screeningData.AML[screeningKey]; exists {
			profileID := s.extractProfileIDFromScreeningResponse(existingData)
			return existingData.TransactionID, profileID, nil
		}
	}

	// merchantID empty, then no data yet, we should get from screening
	if merchantID == "" {
		screeningResponse, err := s.Screening(ctx, request, provider, merchantID)
		if err != nil {
			return "", "", err
		}

		profileID := s.extractProfileIDFromScreeningResponse(screeningResponse)
		return screeningResponse.TransactionID, profileID, nil
	}

	return "", "", constant.ErrDataNotFound
}

// extractProfileIDFromScreeningResponse extracts profileID from ScreeningResponse
func (s *AmlService) extractProfileIDFromScreeningResponse(screeningResponse *amlcommon.ScreeningResponse) string {
	if screeningResponse.Result != nil && len(screeningResponse.Result.Detail) > 0 {
		if screeningResponse.Result.Detail[0].ProfileID != "" {
			return screeningResponse.Result.Detail[0].ProfileID
		}
	}
	return ""
}
