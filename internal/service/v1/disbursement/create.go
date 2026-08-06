package disbursementService

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *DisbursementService) CreateSingle(ctx context.Context, request *disbursementModel.CreateSingleRequest) (*disbursementModel.Disbursement, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/CreateSingle")
	defer segment.End()

	var (
		disbursementID       = uuid.NewString()
		disbursementCurrency = "IDR"                    // This is not constant, will be replacing with XB if any.
		amount, _            = request.Amount.Float64() // Passing this value to snap core, not total amount
		accountInquiryId     *string
		err                  error
		trxFeeOnBehalf       *feeModel.TrxFeeOnBehalfMetadata
	)

	// Get Merchant Data From Context
	merchant, ok := ctx.Value(constant.CtxMerchantData).(*merchantModel.Merchant)
	if !ok {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	merchantId := request.MerchantID
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		merchantId = parentMerchantId
	}

	bankChannelCode := ""
	if bank := bankTransfer.NewBankDB().FindByCode(request.BeneficiaryBankCode); bank != nil {
		bankChannelCode = bank.ChannelCode
	}

	// Check if beneficiary is a Virtual Account
	isVirtualAccount := false
	beneficiaryAccount, err := s.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &beneficiaryAccountModel.CheckAccountRequest{
		MerchantID:           merchantId,
		BeneficiaryBankCode:  request.BeneficiaryBankCode,
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
	})
	if err != nil {
		if request.CreatedFrom != constant.DisbursementCreatedFromOpenApi {
			return nil, err
		}

		// NOTE: Account inquiry is duplicated in the approval/processing flow at
		// process.go:439 (ValidateBankAccountAndUpdateTransaction). An inquiry failure here is non-fatal for OPEN API
		s.logger.Error(ctx, "[create-disbursement] account inquiry failed at creation, deferring rejection to processing", logger.Error(err))
	}
	if beneficiaryAccount != nil {
		isVirtualAccount = beneficiaryAccount.MetadataObj.IsVirtualAccount
	}

	feeReference := constant.ReferenceDisbursement
	if isVirtualAccount {
		// Determine fee reference based on VA status
		feeReference = constant.ReferenceDisbursementVA
		// Check if the payout destination is a Pivot internal VA. If it matches the internal VA pattern, the transaction is declined.
		if !constant.IsPayoutToVirtualAccountAllowed(beneficiaryAccount.BeneficiaryBankCode, beneficiaryAccount.BeneficiaryAccountNo) {
			return nil, pkgErrors.New(response.HttpErrRequest, errors.New(constant.ErrDetailMsgPayoutDstNotEligible))
		}
	}

	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		trxFeeOnBehalf, err = s.feeSvc.GetTransactionFeeOnBehalf(
			ctx, &feeModel.GetTrxFeeOnBehalfRequest{
				MerchantId:        parentMerchantId,
				SubMerchantId:     request.MerchantID,
				Reference:         feeReference,
				TransactionAmount: amount,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	feeAmount, feeDetail, err := s.feeSvc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
		MerchantID:      merchantId,
		Reference:       feeReference,
		Channel:         bankChannelCode,
		ReferenceAmount: amount,
	})
	if err != nil {
		return nil, err
	}
	metadata := disbursementModel.Metadata{
		FeeDetail:   *feeDetail,
		FeeOnBehalf: trxFeeOnBehalf,
	}
	if trxFeeOnBehalf != nil {
		feeAmount = trxFeeOnBehalf.FinalAmount
	}
	totalAmount := amount + feeAmount

	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		metadata.OnBehalf = &merchantModel.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}

	if request.InquiryID != "" {
		accountInquiryId = &request.InquiryID
	}

	derivedMerchantName := request.MerchantName
	if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		if parentMerchant, errParent := s.merchantRepo.FindMerchantByID(ctx, merchant.ParentID.String); errParent != nil {
			return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrFindMerchant)
		} else if parentMerchant != nil {
			derivedMerchantName = parentMerchant.Name
		}
	}

	fee := decimal.NewFromFloat(feeAmount)
	disbursement := &disbursementModel.Disbursement{
		UUID:                   disbursementID,
		ReferenceID:            request.ReferenceID,
		MerchantID:             request.MerchantID,
		BulkID:                 request.BulkID,
		PurposeID:              &request.PurposeID,
		SenderName:             derivedMerchantName,
		AccountInquiryID:       accountInquiryId,
		BeneficiaryBankCode:    request.BeneficiaryBankCode,
		BeneficiaryBankName:    &request.BeneficiaryBankName,
		BeneficiaryAccountNo:   request.BeneficiaryAccountNo,
		BeneficiaryAccountName: request.BeneficiaryAccountName,
		ProcessorReferenceID:   nil,
		Currency:               disbursementCurrency,
		Amount:                 decimal.NewFromFloat(amount),
		Fee:                    &fee,
		TotalAmount:            decimal.NewFromFloat(totalAmount),
		Status:                 constant.DisbursementStatusWaiting,
		ReasonType:             nil,
		ReasonDescription:      nil,
		Remark:                 &request.Remark,
		CreatedFrom:            &request.CreatedFrom,
		CreatedBy:              request.CreatedBy,
		ApprovedBy:             nil,
		ApprovedAt:             nil,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
		MetadataObj:            metadata,
	}
	disbursement.Metadata.Valid = true
	disbursement.Metadata.JSONText, _ = json.Marshal(metadata)

	if err = s.disbursementRepo.Insert(ctx, disbursement); err != nil {
		return nil, err
	}

	createdBy := constant.UserSystemType
	if request.CreatedBy != nil {
		createdBy = *request.CreatedBy
	}
	// Record status history for disbursement creation
	s.recordDisbursementWaiting(ctx, disbursementID, createdBy)

	return disbursement, nil
}
