package accountinquiry

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *AccountInquiryService) CheckStatusRequestInquiry(ctx context.Context, merchantID, inquiryID string) (*requestAccountInquiries.RequestAccountInquiriesHttpResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/requestAccountInquiry/isRequestStatusIsPending")
	defer span.End()

	var (
		status string
	)

	requestInquiry, err := s.repo.FindLatestByInquiryID(ctx, inquiryID, merchantID)
	if err != nil {
		s.logger.Error(ctx, "error when FindLatestByInquiryID", logger.Error(err))
		return nil, err
	}

	if requestInquiry == nil {
		s.logger.Warn(ctx, "request inquiry not found", logger.String("inquiry_id", inquiryID), logger.String("merchant_id", merchantID))
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrInquiryIdNotFound)
	}

	_ = json.Unmarshal(requestInquiry.Metadata.JSONText, &requestInquiry.MetadataObj)

	if requestInquiry.Status.String != constant.RequestAccountInquiryStatusPending {
		return &requestAccountInquiries.RequestAccountInquiriesHttpResponse{
			UUID:       inquiryID,
			MerchantID: requestInquiry.MerchantID,
			InquiryResult: requestAccountInquiries.InquiryResult{
				Status: requestInquiry.Status.String,
				Detail: requestInquiry.MetadataObj.DetailStatus,
			},
			AdditionalInfo: requestAccountInquiries.BuildAccountInquiryAdditionalInfo(
				requestInquiry.MerchantID,
				&requestInquiry.MetadataObj,
				nil,
			),
		}, nil
	}

	payload := requestAccountInquiries.RequestAccountInquiriesHttpRequest{
		RequestInquiryID: requestInquiry.UUID,
		MerchantID:       requestInquiry.MerchantID,
		ChannelCode:      requestInquiry.BeneficiaryBankCode,
		ChannelInformation: requestAccountInquiries.ChannelInformation{
			AccountName:   requestInquiry.BeneficiaryAccountName.String,
			AccountNumber: requestInquiry.BeneficiaryAccountNo.String,
		},
	}

	inquiry, err := s.processInquiryToProcessor(ctx, payload)
	if err != nil || inquiry.BeneficiaryAccountName == "" {
		status = constant.RequestAccountInquiryStatusInvalid
	} else if util.IsPatternMatch(constant.SnapCoreResponseCodeRequestInProgress, inquiry.ResponseCode) {
		status = constant.RequestAccountInquiryStatusPending
	} else if strings.ReplaceAll(strings.ToUpper(inquiry.BeneficiaryAccountName), " ", "") != strings.ReplaceAll(strings.ToUpper(payload.ChannelInformation.AccountName), " ", "") {
		status = constant.RequestAccountInquiryStatusWarning
	} else {
		status = constant.RequestAccountInquiryStatusValid
	}

	_, detailStatus := requestAccountInquiries.NewDetailStatusRequestInquiry(status, inquiry.BeneficiaryAccountName, payload.ChannelInformation.AccountName, "", inquiry.ResponseCode)

	if status == constant.RequestAccountInquiryStatusPending {
		return &requestAccountInquiries.RequestAccountInquiriesHttpResponse{
			UUID:       inquiryID,
			MerchantID: requestInquiry.MerchantID,
			InquiryResult: requestAccountInquiries.InquiryResult{
				Status: status,
				Detail: detailStatus,
			},
			AdditionalInfo: requestAccountInquiries.BuildAccountInquiryAdditionalInfo(
				requestInquiry.MerchantID,
				&requestInquiry.MetadataObj,
				inquiry,
			),
		}, nil
	}

	ctx, err = s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "error when begin transaction", logger.Error(err))
		return nil, err
	}

	processDone := false
	defer func() {
		if !processDone {
			err = s.repo.RollbackTransaction(ctx)
			if err != nil {
				s.logger.Error(ctx, "error when rollback transaction", logger.Error(err))
			}
		}
	}()

	requestInquiry.Status = sql.NullString{
		String: status,
		Valid:  true,
	}

	requestInquiry.MetadataObj.DetailStatus = detailStatus
	requestInquiry.MetadataObj.SnapCoreResponse = inquiry
	requestInquiry.BeneficiaryAccountName = sql.NullString{
		String: inquiry.BeneficiaryAccountName,
		Valid:  true,
	}
	requestInquiry.BeneficiaryBankName = sql.NullString{
		String: inquiry.BeneficiaryBankName,
		Valid:  true,
	}

	requestInquiry.SetMetadataNullJSONText()
	if err := s.repo.Update(ctx, requestInquiry); err != nil {
		s.logger.Error(ctx, "error when update request inquiry", logger.Error(err))
		return nil, err
	}

	beneficiaryAccount, err := s.beneficiaryRepo.GetByID(ctx, requestInquiry.AccountInquiryId.String)
	if err != nil {
		s.logger.Error(ctx, "error when get account inquiry", logger.Error(err))
		return nil, err
	}

	if beneficiaryAccount == nil {
		s.logger.Warn(ctx, "account inquiry not found", logger.String("account_inquiry_id", requestInquiry.AccountInquiryId.String))
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrAccountNotFound)
	}

	if status != constant.RequestAccountInquiryStatusInvalid {
		beneficiaryAccount.MetadataObj.RequestInquiryStatus = status
		beneficiaryAccount.MetadataObj.IsVirtualAccount = inquiry.IsVirtualAccount

		beneficiaryAccount.BeneficiaryAccountName = inquiry.BeneficiaryAccountName
		beneficiaryAccount.BeneficiaryAccountNo = inquiry.BeneficiaryAccountNo
		beneficiaryAccount.BeneficiaryBankCode = inquiry.BeneficiaryBankCode
		beneficiaryAccount.BeneficiaryBankName = inquiry.BeneficiaryBankName
		beneficiaryAccount.Metadata.Valid = true
		beneficiaryAccount.Metadata.JSONText, _ = json.Marshal(beneficiaryAccount.MetadataObj)
		beneficiaryAccount.UpdatedAt = time.Now().UTC()

		err := s.beneficiaryRepo.Update(ctx, beneficiaryAccount)
		if err != nil {
			s.logger.Error(ctx, "error when update account inquiry", logger.Error(err))
			return nil, err
		}
	}

	err = s.repo.CommitTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "error when commit transaction", logger.Error(err))
		return nil, err
	}
	processDone = true

	return &requestAccountInquiries.RequestAccountInquiriesHttpResponse{
		UUID:       inquiryID,
		MerchantID: requestInquiry.MerchantID,
		InquiryResult: requestAccountInquiries.InquiryResult{
			Status: status,
			Detail: detailStatus,
		},
		AdditionalInfo: requestAccountInquiries.BuildAccountInquiryAdditionalInfo(
			requestInquiry.MerchantID,
			&requestInquiry.MetadataObj,
			inquiry,
		),
	}, nil
}
