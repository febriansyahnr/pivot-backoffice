package dukcapilservice

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/dukcapil"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DukcapilService) VerifyIdentity(ctx context.Context, request *dukcapilmodel.IdentityVerificationRequest) (*dukcapilmodel.IdentityVerificationResponse, error) {
	ctx, segment := tracer.Start(ctx, "/internal/service/v1/dukcapil/verifyIdentity")
	defer segment.End()

	referenceID, _ := uuid.NewV7()

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		ReferenceId: referenceID.String(),
		RequestId:   traceID,
		From:        constant.DUKCAPIL_GATEWAY,
	})

	verificationResult, err := s.dukcapilGatewayRepo.VerifyIdentity(ctx, request.VerifyRequest)
	if err != nil {
		s.logger.Error(ctx, "error calling Dukcapil gateway", logger.Error(err))
		return nil, err
	}

	fieldResults := s.validateFields(ctx, verificationResult)
	overallStatus := s.determineOverallStatus(fieldResults)

	response := &dukcapilmodel.IdentityVerificationResponse{
		ReferenceID:  referenceID.String(),
		Status:       overallStatus,
		FieldResults: fieldResults,
	}

	// only save to merchant if provided
	if request.MerchantID != "" {
		err = s.storeVerificationResult(ctx, request.MerchantID, response, request.VerifyRequest, verificationResult)
		if err != nil {
			s.logger.Error(ctx, "error storing verification result", logger.Error(err))
			return nil, err
		}
	}

	return response, nil
}

func (s *DukcapilService) validateFields(ctx context.Context, result *dukcapilmodel.VerifyResult) []dukcapilmodel.DukcapilFieldResult {
	var fieldResults []dukcapilmodel.DukcapilFieldResult

	fieldThresholds := s.GetFieldThresholds()

	fieldMappings := dukcapilmodel.NewDukcapilFieldMappings(result)

	for _, fieldMapping := range fieldMappings.Fields {
		threshold := dukcapilmodel.GetThresholdForField(fieldThresholds, fieldMapping.StandardField)
		score := dukcapilmodel.ParseDukcapilFieldScore(fieldMapping.Value)

		status := dukcapilmodel.StatusNotMatched
		if score >= threshold {
			status = dukcapilmodel.StatusMatched
		}

		fieldResults = append(fieldResults, dukcapilmodel.DukcapilFieldResult{
			Field:     fieldMapping.StandardField,
			Score:     score,
			Threshold: threshold,
			Status:    status,
		})
	}

	return fieldResults
}

// naive status matching, if one of the field score is below threshold
// then return NOT_MATCHED
func (s *DukcapilService) determineOverallStatus(fieldResults []dukcapilmodel.DukcapilFieldResult) string {
	for _, result := range fieldResults {
		if result.Status == dukcapilmodel.StatusNotMatched {
			return dukcapilmodel.StatusNotMatched
		}
	}
	return dukcapilmodel.StatusMatched
}

func (s *DukcapilService) storeVerificationResult(ctx context.Context, merchantID string, response *dukcapilmodel.IdentityVerificationResponse, request *dukcapilmodel.VerifyRequest, result *dukcapilmodel.VerifyResult) error {
	ctx, segment := tracer.Start(ctx, "/internal/service/v1/dukcapil/storeVerificationResult")
	defer segment.End()

	if request.Name == "" {
		s.logger.Error(ctx, "person name is required for screening data key")
		return constant.ErrInvalidRequestPayload
	}
	if request.DOB == "" {
		s.logger.Error(ctx, "date of birth is required for screening data key")
		return constant.ErrInvalidRequestPayload
	}

	screeningKey := request.Name + ":" + request.DOB

	merchant, err := s.merchantRepository.FindMerchantByID(ctx, merchantID)
	if err != nil || merchant == nil {
		s.logger.Error(ctx, "merchant not found", logger.Error(err), logger.String("merchantID", merchantID))
		return constant.ErrMerchantNotFound
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

	if screeningData.Dukcapil == nil {
		screeningData.Dukcapil = make(map[string]*dukcapilmodel.DukcapilScreeningData)
	}

	if existingData, exists := screeningData.Dukcapil[screeningKey]; exists {
		s.logger.Info(ctx, "dukcapil screening data already exists for person, updating",
			logger.String("screeningKey", screeningKey),
			logger.String("existingStatus", existingData.Status),
			logger.String("newStatus", response.Status))
	}

	screeningData.Dukcapil[screeningKey] = &dukcapilmodel.DukcapilScreeningData{
		Status:       response.Status,
		FieldResults: response.FieldResults,
	}

	updatedData, err := json.Marshal(screeningData)
	if err != nil {
		s.logger.Error(ctx, "failed to marshal screening data", logger.Error(err))
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

	s.logger.Info(ctx, "successfully saved dukcapil screening data to merchant",
		logger.String("merchantID", merchantID),
		logger.String("screeningKey", screeningKey),
		logger.String("status", response.Status))

	return nil
}
