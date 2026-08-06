package beneficiaryAccountService

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/go/snap/bankTransfer"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *BeneficiaryAccountService) getWhitelistAccountNo(ctx context.Context, accountNo string) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/BeneficiaryAccountService/getWhitelistAccountNo")
	defer segment.End()

	accountNoFlag := ffcontext.NewEvaluationContext(accountNo)
	accountNoFlag.AddCustomAttribute("account_no", accountNo)

	whitelistIsTrue, err := ffclient.BoolVariation("backend-portal-beneficiaryAccount-accountNo-bypass-whitelist", accountNoFlag, false)
	if err != nil {
		s.logger.Debug(ctx, "error when try to feature flag bool variation backend-portal-beneficiaryAccount-accountNo-bypass-whitelist", logger.Error(err))
		return false, err
	}

	return whitelistIsTrue, nil
}

func (s *BeneficiaryAccountService) FindByBankCodeAndAccountNo(
	ctx context.Context,
	req *beneficiaryAccountModel.CheckAccountRequest,
) (*beneficiaryAccountModel.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/BeneficiaryAccountService/FindByBankCodeAndAccountNo")
	defer segment.End()

	var (
		beneficiaryAccountMetadata     types.NullJSONText
		updatedBeneficiaryMetadata     beneficiaryAccountModel.Metadata
		isUpdateDataBeneficiaryAccount bool
	)

	// Check if the account no is whitelisted for ALWAYS checking to snap core
	// For PAT purposes, so we do not have to delete every time in database
	alwaysCheckSnapCore, _ := s.getWhitelistAccountNo(ctx, req.BeneficiaryAccountNo)
	if alwaysCheckSnapCore {
		// if the account is not existed in both beneficiary_account and account_inquiries
		newReq := &routingProcessorModel.InquiryAccountRequest{
			BeneficiaryBankCode:  req.BeneficiaryBankCode,
			BeneficiaryAccountNo: req.BeneficiaryAccountNo,
			AdditionalInfo:       req.AdditionalInfo,
			MerchantID:           req.MerchantID,
		}

		ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
			ReferenceId: req.MerchantID,
			From:        "Beneficiary-Account-Service",
		})

		// call snapCore
		snapCoreResp, err := s.routingProcessorSvc.AccountInquiry(ctx, newReq)
		if err != nil {
			if snapCoreResp != nil {
				snapRespCode := snapCoreResp.ResponseCode
				errDetails := util.MapAccountInquirySnapResponseToDetailStatus(snapRespCode)

				return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, errors.New(errDetails))
			}

			return nil, err
		}

		// convert to account
		account := &beneficiaryAccountModel.Account{
			UUID:                   "05b146b2-7241-4d0a-a765-88thisismock",
			MerchantID:             req.MerchantID,
			BeneficiaryAccountNo:   req.BeneficiaryAccountNo,
			BeneficiaryAccountName: snapCoreResp.BeneficiaryAccountName,
			BeneficiaryBankCode:    req.BeneficiaryBankCode,
			BeneficiaryBankName:    snapCoreResp.BeneficiaryBankName,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
			MetadataObj: beneficiaryAccountModel.Metadata{
				IsVirtualAccount: snapCoreResp.IsVirtualAccount,
			},
		}

		return account, nil
	}

	merchantID := req.MerchantID
	// when the merchant was non-kyc account, then should use parent merchant for validation
	if derivedID, ok := ctx.Value(constant.CtxDerivedMerchantID).(string); ok {
		merchantID = derivedID
	}

	// check if the account is existed in beneficiary_account
	beneficiaryAccount, err := s.beneficiaryAccountRepo.GetByBankCodeAndAccountNo(ctx, merchantID, req.BeneficiaryBankCode, req.BeneficiaryAccountNo)
	if err != nil {
		return nil, err
	}

	// if the account is existed in beneficiary_account
	if beneficiaryAccount != nil {
		// convert to account
		account := &beneficiaryAccountModel.Account{
			UUID:                   beneficiaryAccount.UUID,
			MerchantID:             beneficiaryAccount.MerchantID,
			BeneficiaryAccountNo:   beneficiaryAccount.BeneficiaryAccountNo,
			BeneficiaryAccountName: beneficiaryAccount.BeneficiaryAccountName,
			BeneficiaryBankCode:    beneficiaryAccount.BeneficiaryBankCode,
			BeneficiaryBankName:    beneficiaryAccount.BeneficiaryBankName,
			CreatedAt:              beneficiaryAccount.CreatedAt,
			UpdatedAt:              beneficiaryAccount.UpdatedAt,
			MetadataObj:            beneficiaryAccount.MetadataObj,
		}

		if account.BeneficiaryAccountName != "" &&
			beneficiaryAccount.MetadataObj.RequestInquiryStatus == constant.RequestAccountInquiryStatusValid {
			return account, nil
		} else {
			// when beneficiary account name is empty, we need re-call processor inquiry account again
			isUpdateDataBeneficiaryAccount = true
		}
	}

	// if the account is not existed in both beneficiary_account and account_inquiries
	newReq := &routingProcessorModel.InquiryAccountRequest{
		BeneficiaryBankCode:  req.BeneficiaryBankCode,
		BeneficiaryAccountNo: req.BeneficiaryAccountNo,
		AdditionalInfo:       req.AdditionalInfo,
		MerchantID:           req.MerchantID,
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		ReferenceId: req.MerchantID,
		From:        "Beneficiary-Account-Service",
	})

	// call snapCore
	snapCoreResp, err := s.routingProcessorSvc.AccountInquiry(ctx, newReq)
	if snapCoreResp != nil {
		if err != nil {
			snapRespCode := snapCoreResp.ResponseCode
			errDetails := util.MapAccountInquirySnapResponseToDetailStatus(snapRespCode)

			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, errors.New(errDetails))
		}

		if util.IsPatternMatch(constant.SnapCoreResponseCodeSuccessPattern, snapCoreResp.ResponseCode) {
			if beneficiaryAccount != nil {
				updatedBeneficiaryMetadata = beneficiaryAccount.MetadataObj
			}
			updatedBeneficiaryMetadata.RequestInquiryStatus = constant.RequestAccountInquiryStatusValid
			updatedBeneficiaryMetadata.IsVirtualAccount = snapCoreResp.IsVirtualAccount

			beneficiaryAccountMetadata.Valid = true
			beneficiaryAccountMetadata.JSONText, _ = json.Marshal(updatedBeneficiaryMetadata)
		}
	}

	if err != nil {
		return nil, err
	}

	if snapCoreResp == nil {
		s.logger.Warn(ctx, "Empty response inquiry from processor", logger.String("accountNo", req.BeneficiaryAccountNo))
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrEmptyProcessorResponse)
	}

	// Override empty beneficiary bank name
	if snapCoreResp.BeneficiaryBankName == "" {
		bankDB := bankTransfer.NewBankDB()
		bank := bankDB.FindByCode(req.BeneficiaryBankCode)
		if bank != nil {
			snapCoreResp.BeneficiaryBankName = bank.Name
		}
	}

	if isUpdateDataBeneficiaryAccount {
		beneficiaryAccount.BeneficiaryAccountName = snapCoreResp.BeneficiaryAccountName
		beneficiaryAccount.BeneficiaryAccountNo = snapCoreResp.BeneficiaryAccountNo
		beneficiaryAccount.BeneficiaryBankCode = snapCoreResp.BeneficiaryBankCode
		beneficiaryAccount.BeneficiaryBankName = snapCoreResp.BeneficiaryBankName
		beneficiaryAccount.UpdatedAt = time.Now().UTC()
		if beneficiaryAccountMetadata.Valid {
			beneficiaryAccount.Metadata = beneficiaryAccountMetadata
			beneficiaryAccount.MetadataObj = updatedBeneficiaryMetadata
		}

		err = s.beneficiaryAccountRepo.Update(ctx, beneficiaryAccount)
		if err != nil && err != constant.ErrNoRowsAffected {
			return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
		}
	} else {
		// save to beneficiary_account
		beneficiaryAccount = &beneficiaryAccountModel.BeneficiaryAccount{
			UUID:                   uuid.NewString(),
			BeneficiaryAccountNo:   snapCoreResp.BeneficiaryAccountNo,
			BeneficiaryAccountName: snapCoreResp.BeneficiaryAccountName,
			BeneficiaryBankCode:    req.BeneficiaryBankCode,
			BeneficiaryBankName:    snapCoreResp.BeneficiaryBankName,
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			MerchantID:             merchantID,
		}

		if beneficiaryAccountMetadata.Valid {
			beneficiaryAccount.Metadata = beneficiaryAccountMetadata
			beneficiaryAccount.MetadataObj = updatedBeneficiaryMetadata
		}

		err = s.beneficiaryAccountRepo.Create(ctx, beneficiaryAccount)
		if err != nil {
			return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
		}
	}

	// convert to account
	account := &beneficiaryAccountModel.Account{
		UUID:                   beneficiaryAccount.UUID,
		MerchantID:             beneficiaryAccount.MerchantID,
		BeneficiaryAccountNo:   beneficiaryAccount.BeneficiaryAccountNo,
		BeneficiaryAccountName: beneficiaryAccount.BeneficiaryAccountName,
		BeneficiaryBankCode:    beneficiaryAccount.BeneficiaryBankCode,
		BeneficiaryBankName:    beneficiaryAccount.BeneficiaryBankName,
		CreatedAt:              beneficiaryAccount.CreatedAt,
		UpdatedAt:              beneficiaryAccount.UpdatedAt,
		MetadataObj:            beneficiaryAccount.MetadataObj,
	}

	return account, nil
}
