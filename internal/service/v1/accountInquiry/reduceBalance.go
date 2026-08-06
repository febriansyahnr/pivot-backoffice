package accountinquiry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

func (s *AccountInquiryService) ReduceBalance(ctx context.Context, requestInquiryID, merchantID string, metadata *requestAccountInquiries.Metadata) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/accountInquiry/ReduceBalance")
	defer segment.End()

	feeAmount := metadata.FeeDetail.FinalAmount
	if metadata.FeeOnBehalf != nil {
		feeAmount = metadata.FeeOnBehalf.FinalAmount
	}

	availableBalance, err := s.orchestratorService.GetAvailableMerchantBalance(ctx, merchantID, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(ctx, "error when get merchant balance", logger.Error(err))
		return constant.ErrValidateBalance

	} else if feeAmount > availableBalance {
		return constant.ErrInsufficientBalance
	}

	merchantUUID, _ := uuid.Parse(merchantID) // Merchant Id or Main Merchant Id (Platform)
	if metadata.FeeOnBehalf != nil {
		if metadata.FeeOnBehalf.FinalAmount > 0 {
			transferRequest := &transfer.TransferRequest{
				SourceMerchantID: util.ParseUUID(merchantID),         // Sub-Merchant
				RecipientID:      metadata.OnBehalf.ParentMerchantId, // Main-Merchant
				ReferenceID:      requestInquiryID,
				TransferType:     constant.MoneyFlowDirect,
				Amount:           metadata.FeeOnBehalf.FinalAmount,
				Remarks:          fmt.Sprintf("Transfer of account inquiry fee - ID: %s", requestInquiryID),
				ParentMerchantID: util.ParseUUID(metadata.OnBehalf.ParentMerchantId), // Main-Merchant
				Usecase:          constant.TypeDisbursement,
			}
			transferResult, err := s.transferSvc.Transfer(ctx, transferRequest)
			if err != nil {
				s.logger.Error(ctx, "RequestAccountInquiry - error when transfer of account inquiry fee", logger.Error(err))
				return pkgErrs.New(response.HttpErrDatabase, err)
			}

			metadata.FeeDetail.TransferId = transferResult.UUID.String()
		}

		metadata.FeeDetail.Notes = "ON-BEHALF"
		merchantUUID = util.ParseUUID(metadata.OnBehalf.ParentMerchantId)
	}

	trxStatus := constant.StatusSuccess
	if metadata.FeeDetail.DeductionType == constant.MerchantFeeDeductionTypeAutomated {
		trxStatus = constant.StatusPending
	}

	feeMetadata, _ := json.Marshal(metadata.FeeDetail)

	err = s.orchestratorService.PostAccountTransaction(ctx, &orchestrator_model.CreateAccountTransactionRequest{
		UUID:                 uuid.New(),
		MerchantID:           merchantUUID,
		Currency:             constant.CurrencyIDR,
		Debit:                metadata.FeeDetail.FinalAmount,
		TransactionTimestamp: time.Now().UTC(),
		ReferenceID:          requestInquiryID,
		Type:                 constant.TypeFee,
		Status:               trxStatus,
		Usecase:              constant.TypeDisbursement,
		AdditionalInfo: types.NullJSONText{
			Valid: true, JSONText: feeMetadata,
		},
	})
	if err != nil {
		s.logger.Error(ctx, "error when reduce merchant balance", logger.Error(err))
		return constant.ErrReduceBalance
	}
	return nil
}
