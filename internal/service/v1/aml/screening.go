package amlservice

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *AmlService) Screening(ctx context.Context, request *amlcommon.CheckRequest, provider string, merchantID string) (*amlcommon.ScreeningResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/aml/Check")
	defer span.End()

	referenceID, _ := uuid.NewV7()

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		ReferenceId: referenceID.String(),
		RequestId:   traceID,
		From:        constant.AML_PROCESSOR,
	})

	response := amlcommon.ScreeningResponse{
		Status: constant.AML_STATUS_APPROVED,
	}

	repo, ok := s.thirdPartyProcessor[provider]
	if !ok {
		err := constant.ErrProviderNotFound
		s.logger.Error(ctx, "provider not found", logger.Error(err), logger.String("provider", provider))
		return &response, err
	}

	request.ReferenceID = referenceID.String()
	response.ReferenceID = referenceID.String()

	checkResponse, err := repo.Check(ctx, request)
	if err != nil {
		s.logger.Error(ctx, constant.ErrAmlCheck.Error(), logger.Error(err))
		return &response, err
	}

	if checkResponse.TransactionID == "" {
		s.logger.Error(ctx, constant.ErrAmlTransactionIDNotExist.Error(), logger.Error(err))
		return &response, err
	}

	response.TransactionID = checkResponse.TransactionID

	inquiryResponse, err := repo.Inquiry(ctx, checkResponse.TransactionID)
	if err != nil {
		s.logger.Error(ctx, constant.ErrAmlInquiry.Error(), logger.Error(err))
		return &response, err
	}

	screeningResult := s.extractScreeningResult(inquiryResponse)
	response.Result = screeningResult

	if checkResponse.Data.Status == constant.AML_STATUS_REVIEW {
		response.Status = constant.AML_STATUS_NEED_REVIEW
	}

	// Save screening result to merchant's third_party_screening_data
	if merchantID != "" {
		err := s.saveScreeningDataToMerchant(ctx, merchantID, request, &response)
		if err != nil {
			s.logger.Warn(ctx, "failed to save screening data to merchant",
				logger.Error(err),
				logger.String("merchantID", merchantID),
				logger.String("personName", request.Name))
		}
	}

	return &response, nil
}

func (s *AmlService) extractScreeningResult(inquiryResponse *amlcommon.InquiryResponse) *amlcommon.ScreeningResult {
	for _, node := range inquiryResponse.Data.Nodes {
		if node.Type == amlcommon.NodeTypeAMLNameScreening && node.Result != nil {
			details := make([]amlcommon.ScreeningDetailItem, len(node.Result.Detail))
			for i, detail := range node.Result.Detail {
				details[i] = amlcommon.ScreeningDetailItem{
					NodeDetail: detail,
					Status:     amlcommon.DetailStatusPending,
				}
			}

			nodeAttrs := amlcommon.ExtractNodeAttributes(node.Attributes)
			attributes := amlcommon.ScreeningAttributes(nodeAttrs)

			return &amlcommon.ScreeningResult{
				ID:            inquiryResponse.Data.ID,
				CompletedAt:   node.CompletedAt,
				TransactionID: inquiryResponse.TransactionID,
				Detail:        details,
				Summary:       node.Result.Summary,
				MatchedCount:  node.Result.MatchedCount,
				Attributes:    attributes,
			}
		}
	}
	return nil
}

func (s *AmlService) saveScreeningDataToMerchant(ctx context.Context, merchantID string, request *amlcommon.CheckRequest, screeningResponse *amlcommon.ScreeningResponse) error {
	ctx, span := tracer.Start(ctx, "internal/service/v1/aml/saveScreeningDataToMerchant")
	defer span.End()

	if request.Name == "" {
		s.logger.Error(ctx, "person name is required for screening data key")
		return constant.ErrInvalidRequestPayload
	}

	var screeningKey string
	if request.SubjectType == constant.AML_SUBJECT_TYPE_ENTITY {
		screeningKey = request.Name
	} else {
		if request.DOB == "" {
			s.logger.Error(ctx, "date of birth is required for screening data key")
			return constant.ErrInvalidRequestPayload
		}
		screeningKey = request.Name + ":" + request.DOB
	}

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

	if screeningData.AML == nil {
		screeningData.AML = make(map[string]*amlcommon.ScreeningResponse)
	}

	if existingData, exists := screeningData.AML[screeningKey]; exists {
		s.logger.Info(ctx, "aml screening data already exists for person, updating",
			logger.String("screeningKey", screeningKey),
			logger.String("existingTransactionID", existingData.TransactionID),
			logger.String("newTransactionID", screeningResponse.TransactionID))
	}

	screeningData.AML[screeningKey] = screeningResponse

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

	s.logger.Info(ctx, "successfully saved aml screening data to merchant",
		logger.String("merchantID", merchantID),
		logger.String("screeningKey", screeningKey),
		logger.String("transactionID", screeningResponse.TransactionID))

	return nil
}
