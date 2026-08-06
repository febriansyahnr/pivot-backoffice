package xbPayoutService

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *xbPayoutService) CreateSession(ctx context.Context, request *xbModel.CreatePayoutSessionRequest) (*xbModel.CreatePayoutSessionResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/CreateSession")
	defer segment.End()

	var (
		payoutUUID, _ = uuid.NewV7()
	)

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    payoutUUID.String(),
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	// Validate reference ID doesn't already exist
	existingDisbursement, err := s.disbursementRepo.FindByMerchantAndReference(ctx, request.MerchantId, request.ReferenceId)
	if err != nil {
		s.logger.Error(ctx, "CreateSession - Failed to check existing reference", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}
	if existingDisbursement != nil {
		s.logger.Warn(ctx, "CreateSession - Reference ID already exists", logger.String("referenceId", request.ReferenceId), logger.String("merchantId", request.MerchantId))
		return nil, pkgErrors.New(response.HttpErrDupCheck, constant.ErrDuplicateDisbursementReferenceId)
	}

	// 1. Create sender if senderId is empty
	if request.SenderID == "" {
		createSenderReq := request.SenderData
		createSenderReq.MerchantId = request.MerchantId
		sender, err := s.CreateSender(ctx, createSenderReq)
		if err != nil {
			s.logger.Error(ctx, "CreateSession - CreateSender failed", logger.Error(err))
			return nil, err
		}

		if !sender.Created {
			s.logger.Error(ctx, "CreateSession - Sender not created", logger.Any("merchantId", request.MerchantId))
			return nil, pkgErrors.New(response.ErrTypeAPIValidation, constant.ErrWhenCreateSenderData)
		}

		request.SenderID = sender.UUID.String()
	}

	// 2. Create beneficiary if beneficiaryId is empty
	if request.BeneficiaryID == "" {
		createBeneficiaryReq := request.BeneficiaryData
		createBeneficiaryReq.MerchantId = request.MerchantId
		beneficiary, err := s.CreateBeneficiary(ctx, createBeneficiaryReq)
		if err != nil {
			s.logger.Error(ctx, "CreateSession - CreateBeneficiary failed", logger.Error(err))
			return nil, err
		}

		if !beneficiary.Created {
			s.logger.Error(ctx, "CreateSession - Beneficiary not created", logger.Any("merchantId", request.MerchantId))
			return nil, pkgErrors.New(response.ErrTypeAPIValidation, constant.ErrWhenCreateBeneficiaryData)
		}

		request.BeneficiaryID = beneficiary.UUID.String()
	}

	// 3. Create session
	xbPayout, err := s.xbCoreProcessorRepo.CreatePayoutSession(ctx, &xbCoreProcessorModel.CreatePayoutSessionRequest{
		MerchantId:          request.MerchantId,
		ReferenceId:         payoutUUID.String(),
		SourceCurrency:      request.SourceCurrency,
		DestinationCurrency: request.DestinationCurrency,
		DestinationAmount:   request.DestinationAmount,
		RemitterId:          request.SenderID,
		BeneficiaryId:       request.BeneficiaryID,
		StatementNarrative:  request.Remark,
		PurposeCode:         request.PurposeCode,
		RoutingValue:        request.RoutingValue,
		CNAPS:               request.CNAPS,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateSession - CreatePayoutSession failed", logger.Error(err))
		return nil, err
	}

	// Insert into disbursement entity
	feeDetail, err := s.insertToDisbursementEntity(
		ctx,
		request,
		payoutUUID.String(),
		request.SenderID,
		request.BeneficiaryID,
		xbPayout,
	)
	if err != nil {
		s.logger.Error(ctx, "CreateSession - InsertToDisbursementEntity failed", logger.Error(err))
		return nil, err
	}

	fee := decimal.NewFromFloat(0)
	if feeDetail != nil {
		fee = decimal.NewFromFloat(feeDetail.Amount)
	}

	// Record status history
	actor := request.CreatedBy
	if actor == "" {
		actor = constant.UserSystemType
	}

	_ = s.RecordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: payoutUUID.String(),
		Status:         constant.XbStatusCreated,
		Actor:          actor,
	})

	return &xbModel.CreatePayoutSessionResponse{
		Uuid:                payoutUUID.String(),
		MerchantId:          request.MerchantId,
		ReferenceId:         request.ReferenceId,
		SourceCurrency:      request.SourceCurrency,
		DestinationCurrency: request.DestinationCurrency,
		DestinationAmount:   request.DestinationAmount,
		FxRate:              xbPayout.FxRate,
		DestinationFxRate:   xbPayout.DestinationFxRate,
		Fee:                 fee,
		TotalAmount:         xbPayout.TotalAmount,
		Remark:              request.Remark,
		CreatedAt:           xbPayout.CreatedAt,
		ExpiredAt:           xbPayout.ExpiredAt,
		Status:              s.mapStatus(xbPayout.Status),
		SenderId:            xbPayout.SenderId,
		BeneficiaryId:       xbPayout.BeneficiaryId,
		BeneficiaryData: xbModel.BeneficiaryDataResponse{
			Name:               xbPayout.BeneficiaryData.Name,
			Address:            xbPayout.BeneficiaryData.Address,
			City:               xbPayout.BeneficiaryData.City,
			Postcode:           xbPayout.BeneficiaryData.Postcode,
			State:              xbPayout.BeneficiaryData.State,
			CountryCode:        xbPayout.BeneficiaryData.CountryCode,
			CountryName:        xbPayout.BeneficiaryData.CountryName,
			AccountType:        xbPayout.BeneficiaryData.AccountType,
			AccountNumber:      xbPayout.BeneficiaryData.AccountNumber,
			BankName:           xbPayout.BeneficiaryData.BankName,
			BankCode:           xbPayout.BeneficiaryData.BankCode,
			ContactCountryCode: xbPayout.BeneficiaryData.ContactCountryCode,
			ContactNumber:      xbPayout.BeneficiaryData.ContactNumber,
			Email:              xbPayout.BeneficiaryData.Email,
		},
		SenderData: xbModel.SenderDataResponse{
			Name:                 xbPayout.SenderData.Name,
			Address:              xbPayout.SenderData.Address,
			City:                 xbPayout.SenderData.City,
			Postcode:             xbPayout.SenderData.Postcode,
			State:                xbPayout.SenderData.State,
			CountryCode:          xbPayout.SenderData.CountryCode,
			CountryName:          xbPayout.SenderData.CountryName,
			AccountType:          xbPayout.SenderData.AccountType,
			IdentificationType:   xbPayout.SenderData.IdentificationType,
			IdentificationNumber: xbPayout.SenderData.IdentificationNumber,
			BankAccountNumber:    xbPayout.SenderData.BankAccountNumber,
			Dob:                  xbPayout.SenderData.Dob,
			ContactCountryCode:   xbPayout.SenderData.ContactCountryCode,
			ContactNumber:        xbPayout.SenderData.ContactNumber,
			SourceOfIncome:       xbPayout.SenderData.SourceOfIncome,
		},
		RoutingCode:  xbPayout.RoutingCode,
		RoutingValue: xbPayout.RoutingValue,
	}, nil
}

func (s *xbPayoutService) insertToDisbursementEntity(ctx context.Context, request *xbModel.CreatePayoutSessionRequest, payoutId, senderId, beneficiaryId string, xbPayout *xbCoreProcessorModel.CreatePayoutSessionResponseData) (*feeModel.FeeMetadataObject, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/insertToDisbursementEntity")
	defer segment.End()

	initialReasonType := constant.XbDisbursementReasonTypeWaitingForConfirmation
	initialReasonDesc := constant.XbDisbursementReasonDescWaitingForConfirmation

	_, feeDetail, err := s.feeSvc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
		MerchantID: request.MerchantId,
		Reference:  constant.TypeXB,
		Channel:    xbPayout.RoutingCode,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateSession - GetFeeCalculationAndDetail failed", logger.Error(err))
		return nil, err
	}

	feeFxRate := s.config.XbCoreProcessorConfig.BaseUSDRate
	getRateQuery := xbCoreProcessorModel.GetFxRateRequest{
		MerchantId:          request.MerchantId,
		SourceCurrency:      "IDR",
		DestinationCurrency: "USD",
		RequestType:         constant.TypeFee,
	}

	if xbFxRate, err := s.xbCoreProcessorRepo.GetFxRate(ctx, &getRateQuery); err == nil {
		if floatRate, _ := xbFxRate.DestinationFxRate.Float64(); floatRate > 0 {
			feeFxRate = floatRate
		}
	} else {
		s.logger.Error(ctx, "failed to get XB fx rate, use default fx rate instead", logger.Error(err))
	}

	// Calculate final fee amount with FX conversion and ceiling - store this as final amount
	// Fee Final Amount * MarkupFxRate, then round up to nearest whole number
	finalFeeAmount := math.Ceil(feeDetail.Amount * feeFxRate)
	feeDetail.Amount = finalFeeAmount
	feeDetail.FinalAmount = finalFeeAmount

	metadata := disbursementModel.Metadata{
		FeeDetail: *feeDetail,
		XbDetail: &xbModel.XbPayoutMetadata{
			Uuid:                xbPayout.Uuid,
			SenderId:            senderId,
			BeneficiaryId:       beneficiaryId,
			SourceCurrency:      request.SourceCurrency,
			DestinationCurrency: request.DestinationCurrency,
			PurposeCode:         request.PurposeCode,
			FxRate:              xbPayout.FxRate,
			FeeFxRate:           decimal.NewFromFloat(feeFxRate),
			DestinationFxRate:   xbPayout.DestinationFxRate,
			DestinationAmount:   request.DestinationAmount,
			SpreadValue:         xbPayout.SpreadValue,
			SpreadType:          xbPayout.SpreadType,
			SourceAmount:        xbPayout.TotalAmount.Round(2), // XB total Amount source currency (before merchant fee)
			TotalAmount:         xbPayout.TotalAmount.Round(2), // Source amount + fee
			ExpiredAt:           xbPayout.ExpiredAt,
			BeneficiaryData: xbModel.BeneficiaryDataResponse{
				Name:               xbPayout.BeneficiaryData.Name,
				Address:            xbPayout.BeneficiaryData.Address,
				City:               xbPayout.BeneficiaryData.City,
				Postcode:           xbPayout.BeneficiaryData.Postcode,
				State:              xbPayout.BeneficiaryData.State,
				CountryCode:        xbPayout.BeneficiaryData.CountryCode,
				CountryName:        xbPayout.BeneficiaryData.CountryName,
				AccountType:        xbPayout.BeneficiaryData.AccountType,
				AccountNumber:      xbPayout.BeneficiaryData.AccountNumber,
				BankName:           xbPayout.BeneficiaryData.BankName,
				BankCode:           xbPayout.BeneficiaryData.BankCode,
				ContactCountryCode: xbPayout.BeneficiaryData.ContactCountryCode,
				ContactNumber:      xbPayout.BeneficiaryData.ContactNumber,
				Email:              xbPayout.BeneficiaryData.Email,
			},
			SenderData: xbModel.SenderDataResponse{
				Name:                 xbPayout.SenderData.Name,
				Address:              xbPayout.SenderData.Address,
				City:                 xbPayout.SenderData.City,
				Postcode:             xbPayout.SenderData.Postcode,
				State:                xbPayout.SenderData.State,
				CountryCode:          xbPayout.SenderData.CountryCode,
				CountryName:          xbPayout.SenderData.CountryName,
				AccountType:          xbPayout.SenderData.AccountType,
				IdentificationType:   xbPayout.SenderData.IdentificationType,
				IdentificationNumber: xbPayout.SenderData.IdentificationNumber,
				BankAccountNumber:    xbPayout.SenderData.BankAccountNumber,
				Dob:                  xbPayout.SenderData.Dob,
				ContactCountryCode:   xbPayout.SenderData.ContactCountryCode,
				ContactNumber:        xbPayout.SenderData.ContactNumber,
				SourceOfIncome:       xbPayout.SenderData.SourceOfIncome,
			},
			RoutingCode:  xbPayout.RoutingCode,
			RoutingValue: xbPayout.RoutingValue,
		},
	}

	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		metadata.OnBehalf = &merchantModel.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}

	// Build disbursement data
	disbursement := &disbursementModel.Disbursement{
		UUID:                   payoutId,
		ReferenceID:            request.ReferenceId,
		MerchantID:             request.MerchantId,
		BulkID:                 nil,
		PurposeID:              nil,
		SenderName:             request.MerchantName,
		AccountInquiryID:       nil,
		BeneficiaryBankCode:    xbPayout.BeneficiaryData.BankCode,
		BeneficiaryBankName:    &xbPayout.BeneficiaryData.BankName,
		BeneficiaryAccountNo:   xbPayout.BeneficiaryData.AccountNumber,
		BeneficiaryAccountName: xbPayout.BeneficiaryData.Name,
		ProcessorReferenceID:   &xbPayout.AcquirerTransactionId,
		Currency:               request.DestinationCurrency,
		Amount:                 request.DestinationAmount,
		Fee:                    &decimal.Zero,
		TotalAmount:            request.DestinationAmount,
		Status:                 constant.DisbursementStatusWaiting,
		ReasonType:             &initialReasonType,
		ReasonDescription:      &initialReasonDesc,
		Remark:                 &request.Remark,
		CreatedFrom:            &request.CreatedFrom,
		CreatedBy:              &request.CreatedBy,
		ApprovedBy:             nil,
		ApprovedAt:             nil,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
		MetadataObj:            metadata,
	}
	disbursement.Metadata.Valid = true
	disbursement.Metadata.JSONText, _ = json.Marshal(metadata)

	if err := s.disbursementRepo.Insert(ctx, disbursement); err != nil {
		s.logger.Error(ctx, "CreateSession - InsertToDisbursementEntity failed", logger.Error(err))
		if strings.Contains(err.Error(), "1062") && strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "disbursement_unique_reference_per_merchant") {
			return nil, pkgErrors.New(response.HttpErrDupCheck, constant.ErrDuplicateDisbursementReferenceId)
		}
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	return feeDetail, nil
}
