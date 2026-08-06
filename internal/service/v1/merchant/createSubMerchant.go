package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	industryUtil "github.com/paper-indonesia/pivot-backoffice/pkg/util/industry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *MerchantService) CreateSubMerchant(ctx context.Context, request *merchantModel.MerchantRequest) (*merchantModel.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/CreateSubMerchant")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, request.RequesterID)
	if err != nil {
		s.logger.Error(ctx, "error when get merchant by id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	if request.PICInvitation {
		if user, err := s.UserSvc.FindUserByEmail(ctx, request.PICEmail); err != nil {
			s.logger.Error(ctx, "failed while find user by email address (pic email)", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, err)

		} else if user != nil {
			return nil, pkgErrs.New(response.HttpErrConflict, constant.ErrEmailAlreadyRegistered)
		}
	}

	// only sub merchant with KYC state that allowed to create sub merchant
	if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrNotAllowedToCreateSubMerchant)
	}

	if request.DistrictId > 0 {
		if loc, err := s.locationRepo.GetDistrictById(ctx, request.DistrictId); err != nil {
			s.logger.Error(ctx, "error when get district by id", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, err)

		} else if loc == nil {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDistrictNotFound)
		}
	}

	parentProvided := request.ParentIndustry != ""
	childProvided := request.ChildIndustry != ""

	// Return early: if only one field is provided (XOR logic)
	if parentProvided != childProvided {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("parent industry and child industry must be provided together"))
	}

	// Return early: if both provided but invalid combination
	if parentProvided && childProvided {
		if valid, err := s.industrySvc.ValidateIndustry(ctx, request.ParentIndustry, request.ChildIndustry); err != nil {
			return nil, pkgErrs.New(response.HttpErrDatabase, err)
		} else if !valid {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("invalid parent and child industry combination"))
		}
	}

	// Return early: if MCC is invalid
	if request.MCC != "" {
		if valid, err := s.industrySvc.IsValidMCC(ctx, request.MCC); err != nil {
			return nil, pkgErrs.New(response.HttpErrDatabase, err)
		} else if !valid {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("invalid MCC code"))
		}
	}

	// Return early: if MCC doesn't match industry combination
	if request.ParentIndustry != "" && request.ChildIndustry != "" && request.MCC != "" {
		if err := s.industrySvc.ValidateIndustryMCCCombination(ctx, request.ParentIndustry, request.ChildIndustry, request.MCC); err != nil {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, err)
		}
	}

	// Return early: if Digital Status is invalid
	if request.DigitalStatus != "" && !industryUtil.IsValidDigitalStatus(request.DigitalStatus) {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("digital status must be 'Digital' or 'Non-digital'"))
	}

	// Return early: if Country of Entity is invalid
	if request.CountryOfEntity != "" {
		country, err := s.countrySvc.FindByCode(ctx, request.CountryOfEntity)
		if err != nil {
			return nil, err
		}
		if country == nil {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("invalid country code for country of entity"))
		}
		request.CountryOfEntity = country.Code
	}

	bankDetail := bankTransfer.NewBankDB().FindByChannelCode(request.BankAccount.ChannelCode)
	if bankDetail == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank channel code not registered"))
	}
	bankAccountDetail, err := s.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &beneficiaryAccountModel.CheckAccountRequest{
		MerchantID:           request.RequesterID,
		BeneficiaryAccountNo: request.BankAccount.AccountNumber,
		BeneficiaryBankCode:  bankDetail.Code,
	})
	if err != nil {
		if errors.Is(err, constant.ErrInvalidAccount) {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrInvalidAccount)
		}
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	callbackApiKey, _ := util.GenerateRandomString(32)
	wrappedApiKey, err := s.encryption.Encrypt(ctx, vault.EncryptRequest{Plaintext: []byte(callbackApiKey)})
	if err != nil {
		s.logger.Error(ctx, "failed while encrypting callback api key", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}
	newSubMerchant, err := request.NewSubMerchant(&wrappedApiKey.Ciphertext, wrappedApiKey.KeyVersion)
	if err != nil {
		s.logger.Error(ctx, "error when create new submerchant", logger.Error(err), logger.Any("request", request))
		return nil, err
	}

	// Auto-approve sub account KYC for staging environment
	if s.config.Environment == constant.EnvironmentStaging && request.KYCStatus == constant.KYCStatusWaitingForDocument {
		mid, err := s.repo.GenerateNewMID(ctx)
		if err != nil {
			s.logger.Error(ctx, "failed to generate MID", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, err)
		}

		newSubMerchant.Status = constant.MerchantStatusActive
		newSubMerchant.MID = sql.NullString{
			Valid:  true,
			String: mid,
		}
		newSubMerchant.KYCStatus = sql.NullString{
			String: constant.KYCStatusApproved,
			Valid:  true,
		}

		request.KYCStatus = constant.KYCStatusApproved
		request.MerchantStatus = constant.MerchantStatusActive
	}

	err = s.MerchantShortNameValidation(ctx, newSubMerchant.ShortName, request.ParentID)
	if err != nil {
		s.logger.Error(ctx, "merchant short name validation failed", logger.Error(err))
		return nil, err
	}

	ctxTX, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "error when create session transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted := false
	defer func() {
		if isCompleted {
			return
		}
		if e := s.repo.RollbackTransaction(ctxTX); e != nil {
			s.logger.Error(ctx, "error when rollback session transaction", logger.Error(e))
		}
	}()

	if err = s.repo.Create(ctxTX, newSubMerchant); err != nil {
		s.logger.Error(ctx, "error when create submerchant", logger.Error(err), logger.Any("request", newSubMerchant))
		return nil, err
	}

	err = s.accountService.CreateMerchantAccount(ctxTX, newSubMerchant.UUID, constant.UserTypeSubMerchant)
	if err != nil {
		s.logger.Error(ctx, "error when create submerchant accounts", logger.Error(err), logger.Any("request", newSubMerchant))
		return nil, err
	}

	if err = s.CreateMerchantAuth(ctxTX, newSubMerchant.UUID); err != nil {
		s.logger.Error(ctx, "error when create merchant auth", logger.Error(err), logger.Any("request", newSubMerchant))
		return nil, err
	}

	createdBy := request.UserID
	if createdBy == "" {
		createdBy = merchant.Name
	}
	bankAccountId, _ := uuid.NewV7()
	bankAccountReq := &bankAccount.BankAccount{
		UUID:                   bankAccountId.String(),
		MerchantID:             newSubMerchant.UUID,
		BeneficiaryAccountNo:   bankAccountDetail.BeneficiaryAccountNo,
		BeneficiaryAccountName: bankAccountDetail.BeneficiaryAccountName,
		BeneficiaryBankCode:    bankAccountDetail.BeneficiaryBankCode,
		BeneficiaryBankName:    bankAccountDetail.BeneficiaryBankName,
		CreatedBy:              createdBy,
		UpdatedBy:              createdBy,
		CreatedAt:              time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err = s.bankAccountRepo.Create(ctxTX, bankAccountReq); err != nil {
		s.logger.Error(ctx, "error when create bank account", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if request.PICInvitation {
		_, err = s.UserSvc.CreateMerchantUser(ctxTX, &user.MerchantUserRequest{
			Email:        request.PICEmail,
			Name:         request.PICName,
			Role:         constant.RoleAdmin,
			MerchantId:   newSubMerchant.UUID,
			MerchantName: newSubMerchant.Name,
			Invitation:   (request.KYCStatus == constant.KYCStatusNotRequired || request.KYCStatus == constant.KYCStatusApproved),
		})
		if err != nil {
			return nil, err
		}
	}

	if err = s.repo.CommitTransaction(ctxTX); err != nil {
		s.logger.Error(ctx, "error when commit session transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	// Enable all payment methods for approved sub merchants in staging (after transaction commit)
	if s.config.Environment == constant.EnvironmentStaging && newSubMerchant.Status == constant.MerchantStatusActive {
		err = s.EnableAllPaymentMethod(ctx, newSubMerchant)
		if err != nil {
			s.logger.Warn(ctx, "failed to enable all payment methods for sub account in staging", logger.Error(err), logger.String("merchant_id", newSubMerchant.UUID))
		} else {
			s.logger.Info(ctx, "enabled all payment methods for sub account in staging", logger.String("merchant_id", newSubMerchant.UUID))
		}

		// First callback: SUB.ACTIVATION.PENDING (CREATED state)
		pendingCallback := &pb.SubAccountRegistration{
			SubAccountId:        newSubMerchant.UUID,
			SubAccountStatus:    constant.MerchantStatusCreated,
			SubAccountKycStatus: constant.KYCStatusInReview,
			UpdatedAt:           timestamppb.New(time.Now().UTC()),
		}
		pendingCallbackReq := &pb.ProcessCallbackRequest{
			Name:       constant.CallbackNameSubAccountRegistration,
			Event:      fmt.Sprintf(constant.CallbackEventSubAccountRegistrationPattern, constant.CallbackStatusPending),
			MerchantId: request.ParentID,
		}
		pendingCallbackReq.Request, _ = anypb.New(pendingCallback)

		_ = s.rabbitMqExt.PublishMerchantCallback(ctx, pendingCallbackReq)

		// Second callback: SUB.ACTIVATION.APPROVED (ACTIVE state)
		approvedCallback := &pb.SubAccountRegistration{
			SubAccountId:        newSubMerchant.UUID,
			SubAccountStatus:    constant.MerchantStatusActive,
			SubAccountKycStatus: constant.KYCStatusApproved,
			UpdatedAt:           timestamppb.New(time.Now().UTC().Add(1 * time.Second)),
		}
		approvedCallbackReq := &pb.ProcessCallbackRequest{
			Name:       constant.CallbackNameSubAccountRegistration,
			Event:      fmt.Sprintf(constant.CallbackEventSubAccountRegistrationPattern, constant.CallbackStatusApproved),
			MerchantId: request.ParentID,
		}
		approvedCallbackReq.Request, _ = anypb.New(approvedCallback)

		_ = s.rabbitMqExt.PublishMerchantCallback(ctx, approvedCallbackReq)

		s.logger.Info(ctx, "sent auto-approval callbacks for sub account in staging", logger.String("merchant_id", newSubMerchant.UUID))

		// Restore original KYC status in response for staging auto-approved submerchants
		// The merchant is approved internally (in DB) but response should still show WAITING_FOR_DOCUMENT
		// to maintain backward compatibility and avoid regression issues
		if request.SubAccountType == constant.MerchantKYCTypeKYC {
			request.KYCStatus = constant.KYCStatusWaitingForDocument
		}

	}

	newSubMerchant.BankAccount = &merchantModel.MerchantBankAccountResponse{
		ChannelCode:   request.BankAccount.ChannelCode,
		AccountNumber: request.BankAccount.AccountNumber,
		BankName:      bankAccountDetail.BeneficiaryBankName,
		AccountName:   bankAccountDetail.BeneficiaryAccountName,
	}

	return newSubMerchant, nil
}
