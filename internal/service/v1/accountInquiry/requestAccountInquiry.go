package accountinquiry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

func (s *AccountInquiryService) RequestAccountInquiry(ctx context.Context, req requestAccountInquiries.RequestAccountInquiriesHttpRequest) (resp *requestAccountInquiries.RequestAccountInquiriesHttpResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/requestAccountInquiry/RequestAccountInquiry")
	defer segment.End()

	defer func() {
		metricData := monitoring.CustomMetric{
			ComponentName:        constant.ComponentNameInquiryAccount,
			MetricName:           constant.MetricNameInquiryAccount,
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			Attributes: map[string]any{
				"merchantId":          req.ParentMerchantID,
				"channelCode":         req.ChannelCode,
				"onBehalfSubmerchant": req.ParentMerchantID != req.MerchantID,
			},
		}
		if err != nil {
			errType, errDetail := pkgErrors.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			s.logger.Error(ctx, "error when record custom metric for account inquiry", logger.Error(errMetric), logger.Any("metricData", metricData))
		}
	}()
	var (
		onBehalfInfo   *merchant.OnBehalfObject
		trxFeeOnBehalf *feeModel.TrxFeeOnBehalfMetadata
	)

	bank := bankDB.FindByChannelCode(req.ChannelCode)
	if bank == nil {
		return nil, pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("channel code not found"))
	}
	req.ChannelCode = bank.Code

	// 1. Check available balance for fee deduction
	merchantId := req.MerchantID
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		merchantId = parentMerchantId

		trxFeeOnBehalf, err = s.feeSvc.GetTransactionFeeOnBehalf(
			ctx, &feeModel.GetTrxFeeOnBehalfRequest{
				MerchantId:    parentMerchantId,
				SubMerchantId: req.MerchantID,
				Reference:     constant.ReferenceAccountInquiry,
			},
		)
		if err != nil {
			s.logger.Error(ctx, "RequestAccountInquiry - error when get transaction fee on-behalf", logger.Error(err))
			return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)
		}

		onBehalfInfo = &merchant.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}

	// when its not derived then its kyc merchant
	// and should handle their own beneficiary data
	if derivedID, _ := ctx.Value(constant.CtxDerivedMerchantID).(string); derivedID == "" {
		merchantId = req.MerchantID
	}

	feeAmount, feeDetail, err := s.feeSvc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
		MerchantID: merchantId, // Merchant Id or Main Merchant Id (Platform)
		Reference:  constant.ReferenceAccountInquiry,
	})
	if err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when get fee calculation and detail", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}
	if trxFeeOnBehalf != nil {
		feeAmount = trxFeeOnBehalf.FinalAmount
	}

	availableBalance, err := s.orchestratorService.GetAvailableMerchantBalance(ctx, req.MerchantID, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when get available merchant balance", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)

	} else if feeAmount > availableBalance {
		return nil, pkgErrors.New(httpResponse.HttpErrForbidden, constant.ErrInsufficientBalance)
	}

	// 2. Request account inquiry (Snap Core Processor)
	var (
		status, beneficiaryID, detail string
	)
	req.RequestInquiryID = uuid.NewString()

	snapCoreResp, err := s.processInquiryToProcessor(ctx, req)
	if err != nil {
		// Check if the error is a downstream infrastructure error that should be propagated
		errType, _ := pkgErrors.ExtractError(err)
		switch errType {
		case httpResponse.HttpErrRequestTimeout,
			httpResponse.HttpErrBadGateway,
			httpResponse.HttpErrServiceUnavailable,
			httpResponse.HttpErrRequestLimitExceeded,
			httpResponse.HttpErrThirdParty,
			httpResponse.HttpErrInternal:
			return nil, err
		}
		status = constant.RequestAccountInquiryStatusInvalid
	} else if util.IsPatternMatch(constant.SnapCoreResponseCodeRequestInProgress, snapCoreResp.ResponseCode) {
		status = constant.RequestAccountInquiryStatusPending
	} else if snapCoreResp.BeneficiaryAccountName == "" {
		status = constant.RequestAccountInquiryStatusInvalid
	} else if strings.ToUpper(req.ChannelInformation.AccountName) != strings.ToUpper(snapCoreResp.BeneficiaryAccountName) {
		status = constant.RequestAccountInquiryStatusWarning
	} else if constant.IsAccountInquiryMerchantNameUseRuneCheck(merchantId) && !strings.EqualFold(snapCoreResp.BeneficiaryAccountName, req.ChannelInformation.AccountName) {
		status = constant.RequestAccountInquiryStatusWarning
	} else {
		status = constant.RequestAccountInquiryStatusValid
	}

	updatedStatus, detail := requestAccountInquiries.NewDetailStatusRequestInquiry(status, req.ChannelInformation.AccountName, snapCoreResp.BeneficiaryAccountName, merchantId, snapCoreResp.ResponseCode)
	if status != updatedStatus {
		// change status to valid, example John, with Sdr. John
		status = updatedStatus
	}

	metadata := requestAccountInquiries.Metadata{
		SnapCoreResponse: snapCoreResp,
		DetailStatus:     detail,
		OnBehalf:         onBehalfInfo,
		FeeOnBehalf:      trxFeeOnBehalf,
		FeeDetail:        feeDetail,
	}

	// 3. Upsert beneficiary account and inquiry account data
	if status != constant.RequestAccountInquiryStatusInvalid {
		if beneficiaryID, err = s.UpsertAccountInquiriesIntoBeneficiary(ctx, merchantId, snapCoreResp); err != nil {
			s.logger.Error(ctx, "RequestAccountInquiry - error when upsert beneficiary", logger.Error(err))
			// Return immediately when beneficiary error occurs, don't proceed with transaction
			return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)
		}
	}

	isCompleted := false

	ctxTx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when start session transaction", logger.Error(err))
		return nil, err
	}
	defer func() {
		if isCompleted {
			return
		}
		if e := s.repo.RollbackTransaction(ctxTx); e != nil {
			s.logger.Error(ctx, "RequestAccountInquiry - error when rollback transaction", logger.Error(e))
		}
	}()

	// 4. Check available balance and reduce inquiry account fee
	if err := s.ReduceBalance(ctxTx, req.RequestInquiryID, req.MerchantID, &metadata); err != nil {
		if errors.Is(err, constant.ErrInsufficientBalance) {
			return nil, pkgErrors.New(httpResponse.HttpErrForbidden, err)
		}
		return nil, err
	}

	// 5. Create request account inquiry
	if err = s.createRequestInquiries(ctxTx, req.RequestInquiryID, beneficiaryID, status, metadata, req, snapCoreResp); err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when create request inquiries", logger.Error(err))
		return nil, err
	}

	if err := s.repo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry - error when commit transaction", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}
	isCompleted = true

	// 6. Response to client
	response := requestAccountInquiries.RequestAccountInquiriesHttpResponse{
		MerchantID: req.MerchantID,
		InquiryResult: requestAccountInquiries.InquiryResult{
			Status: status,
			Detail: detail,
		},
	}
	if status != constant.RequestAccountInquiryStatusInvalid {
		response.UUID = beneficiaryID
	}
	response.AdditionalInfo = requestAccountInquiries.BuildAccountInquiryAdditionalInfo(req.MerchantID, &metadata, snapCoreResp)
	return &response, nil
}

func (s *AccountInquiryService) UpsertAccountInquiriesIntoBeneficiary(ctx context.Context, merchantID string, snapCoreResp *snapCoreModel.InquiryAccountResponseData) (beneficiaryID string, err error) {
	beneficiary, err := s.beneficiaryRepo.GetByBankCodeAndAccountNo(ctx, merchantID, snapCoreResp.BeneficiaryBankCode, snapCoreResp.BeneficiaryAccountNo)
	if err != nil {
		s.logger.Error(ctx, "UpsertAccountInquiriesIntoBeneficiary - error when getting beneficiary", logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}

	if beneficiary != nil {
		beneficiary.BeneficiaryAccountName = snapCoreResp.BeneficiaryAccountName
		beneficiary.BeneficiaryAccountNo = snapCoreResp.BeneficiaryAccountNo
		beneficiary.BeneficiaryBankCode = snapCoreResp.BeneficiaryBankCode
		beneficiary.BeneficiaryBankName = snapCoreResp.BeneficiaryBankName
		beneficiary.UpdatedAt = time.Now().UTC()

		if util.IsPatternMatch(constant.SnapCoreResponseCodeSuccessPattern, snapCoreResp.ResponseCode) {
			beneficiary.MetadataObj.RequestInquiryStatus = constant.RequestAccountInquiryStatusValid
			beneficiary.MetadataObj.IsVirtualAccount = snapCoreResp.IsVirtualAccount
			beneficiary.Metadata.Valid = true
			beneficiary.Metadata.JSONText, _ = json.Marshal(beneficiary.MetadataObj)
		}

		if err := s.beneficiaryRepo.Update(ctx, beneficiary); err != nil {
			s.logger.Error(ctx, "UpsertAccountInquiriesIntoBeneficiary - error when updating beneficiary", logger.Error(err))
			return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
		}
		return beneficiary.UUID, nil
	}

	beneficiaryID = uuid.NewString()
	newBeneficiary := &beneficiaryAccountModel.BeneficiaryAccount{
		UUID:                   beneficiaryID,
		MerchantID:             merchantID,
		BeneficiaryAccountName: snapCoreResp.BeneficiaryAccountName,
		BeneficiaryAccountNo:   snapCoreResp.BeneficiaryAccountNo,
		BeneficiaryBankCode:    snapCoreResp.BeneficiaryBankCode,
		BeneficiaryBankName:    snapCoreResp.BeneficiaryBankName,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	if util.IsPatternMatch(constant.SnapCoreResponseCodeSuccessPattern, snapCoreResp.ResponseCode) {
		newBeneficiary.MetadataObj.RequestInquiryStatus = constant.RequestAccountInquiryStatusValid
		newBeneficiary.MetadataObj.IsVirtualAccount = snapCoreResp.IsVirtualAccount
		newBeneficiary.Metadata.Valid = true
		newBeneficiary.Metadata.JSONText, _ = json.Marshal(newBeneficiary.MetadataObj)
	}

	if err := s.beneficiaryRepo.Create(ctx, newBeneficiary); err != nil {
		s.logger.Error(ctx, "UpsertAccountInquiriesIntoBeneficiary - error when creating beneficiary", logger.Error(err))
		return "", pkgErrors.New(httpResponse.HttpErrDatabase, err)
	}
	return beneficiaryID, nil
}

func (s *AccountInquiryService) processInquiryToProcessor(ctx context.Context, req requestAccountInquiries.RequestAccountInquiriesHttpRequest) (*snapCoreModel.InquiryAccountResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/requestAccountInquiry/processInquiryToProcessor")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		ReferenceId: req.MerchantID,
		OriginId:    req.RequestInquiryID,
		From:        "Account-Inquiry-Service",
	})

	payload := &routingProcessorModel.InquiryAccountRequest{
		MerchantID:             req.MerchantID,
		BeneficiaryBankCode:    req.ChannelCode,
		BeneficiaryAccountNo:   req.ChannelInformation.AccountNumber,
		BeneficiaryAccountName: req.ChannelInformation.AccountName,
	}

	if req.AdditionalInfo != nil {
		payload.AdditionalInfo = req.AdditionalInfo
	}

	resp, err := s.routingProcessorSvc.AccountInquiry(ctx, payload)
	if err != nil {
		s.logger.Error(ctx, "RequestAccountInquiry/processInquiryToProcessor - error get bank account inquiry", logger.Error(err))
		errResp := &snapCoreModel.InquiryAccountResponseData{}
		if resp != nil {
			errResp = &snapCoreModel.InquiryAccountResponseData{
				ResponseCode:           resp.ResponseCode,
				ResponseMessage:        resp.ResponseMessage,
				PartnerReferenceNo:     resp.PartnerReferenceNo,
				BeneficiaryAccountName: resp.BeneficiaryAccountName,
				BeneficiaryAccountNo:   resp.BeneficiaryAccountNo,
				BeneficiaryBankCode:    resp.BeneficiaryBankCode,
				BeneficiaryBankName:    resp.BeneficiaryBankName,
				IsVirtualAccount:       resp.IsVirtualAccount,
			}
		}

		return errResp, err
	}

	if resp != nil {
		if util.IsPatternMatch(constant.SnapCoreResponseCodeInactiveAccountPattern, resp.ResponseCode) ||
			util.IsPatternMatch(constant.SnapCoreResponseCodeDormantAccountPattern, resp.ResponseCode) ||
			util.IsPatternMatch(constant.SnapCoreResponseCodeInvalidAccountPattern, resp.ResponseCode) ||
			strings.EqualFold(resp.ResponseMessage, "Invalid Account") {
			// this error will handling just for request account inquiry,
			// when the response code is inactive account, dormant account, or invalid account
			// that will do empty beneficiary account name
			resp.BeneficiaryAccountName = ""
		}
	}

	snapCoreResp := &snapCoreModel.InquiryAccountResponseData{
		ResponseCode:           resp.ResponseCode,
		ResponseMessage:        resp.ResponseMessage,
		PartnerReferenceNo:     resp.PartnerReferenceNo,
		BeneficiaryAccountName: resp.BeneficiaryAccountName,
		BeneficiaryAccountNo:   resp.BeneficiaryAccountNo,
		BeneficiaryBankCode:    resp.BeneficiaryBankCode,
		BeneficiaryBankName:    resp.BeneficiaryBankName,
		IsVirtualAccount:       resp.IsVirtualAccount,
	}

	return snapCoreResp, nil
}

func (s *AccountInquiryService) createRequestInquiries(
	ctx context.Context,
	requestInquiryId string,
	inquiryID string,
	status string,
	metadata requestAccountInquiries.Metadata,
	req requestAccountInquiries.RequestAccountInquiriesHttpRequest,
	snapCoreData *snapCoreModel.InquiryAccountResponseData,
) error {

	metadataB, _ := json.Marshal(metadata)

	createRequestInquiry := &requestAccountInquiries.RequestAccountInquiries{
		UUID:                requestInquiryId,
		MerchantID:          req.MerchantID,
		BeneficiaryBankCode: req.ChannelCode,
		AccountInquiryId: sql.NullString{
			String: inquiryID,
			Valid:  true,
		},
		BeneficiaryAccountName: sql.NullString{
			String: req.ChannelInformation.AccountName,
			Valid:  true,
		},
		BeneficiaryAccountNo: sql.NullString{
			String: req.ChannelInformation.AccountNumber,
			Valid:  true,
		},
		Status: sql.NullString{
			String: status,
			Valid:  true,
		},
		Metadata: types.NullJSONText{
			JSONText: metadataB,
			Valid:    true,
		},
		CreatedAt: time.Now().UTC(),
	}

	if snapCoreData != nil {
		createRequestInquiry.BeneficiaryBankName = sql.NullString{
			String: snapCoreData.BeneficiaryBankName,
			Valid:  true,
		}
	}

	return s.repo.Create(ctx, createRequestInquiry)
}
