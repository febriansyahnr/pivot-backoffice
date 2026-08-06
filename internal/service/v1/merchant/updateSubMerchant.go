package merchant

import (
	"context"
	"fmt"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	industry "github.com/paper-indonesia/pivot-backoffice/pkg/util/industry"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) UpdateSubMerchant(ctx context.Context, request *merchantModel.UpdateMerchantRequest) (*merchantModel.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateSubMerchant")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, request.ID)
	if err != nil {
		s.logger.Error(ctx, "error when find merchant by id", logger.Error(err), logger.Any("request", request))
		return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrUpdateMerchant)

	} else if merchant == nil {
		return nil, errors.New(responseHttp.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	needSendLinkInvitation := merchant.PICInvitation == constant.MerchantPICNotInvited && request.PICInvitation
	if needSendLinkInvitation {
		if user, err := s.UserSvc.FindUserByEmail(ctx, request.PICEmail); err != nil {
			s.logger.Error(ctx, "failed while find user by email address (pic email)", logger.Error(err))
			return nil, errors.New(responseHttp.HttpErrDatabase, err)

		} else if user != nil {
			return nil, errors.New(responseHttp.HttpErrConflict, constant.ErrEmailAlreadyRegistered)
		}
	}

	if request.DistrictId > 0 {
		if loc, err := s.locationRepo.GetDistrictById(ctx, request.DistrictId); err != nil {
			s.logger.Error(ctx, "error when get district by id", logger.Error(err))
			return nil, errors.New(responseHttp.HttpErrDatabase, err)

		} else if loc == nil {
			return nil, errors.New(responseHttp.HttpErrUnprocessableContent, constant.ErrDistrictNotFound)
		}
	}

	parentProvided := request.ParentIndustry != ""
	childProvided := request.ChildIndustry != ""

	// Return early: if only one field is provided (XOR logic)
	if parentProvided != childProvided {
		return nil, errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("parent industry and child industry must be provided together"))
	}

	// Return early: if both provided but invalid combination
	if parentProvided && childProvided {
		if valid, err := s.industrySvc.ValidateIndustry(ctx, request.ParentIndustry, request.ChildIndustry); err != nil {
			return nil, errors.New(responseHttp.HttpErrInternal, err)
		} else if !valid {
			return nil, errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid parent and child industry combination"))
		}
	}

	// Return early: if MCC is invalid
	if request.MCC != "" {
		if valid, err := s.industrySvc.IsValidMCC(ctx, request.MCC); err != nil {
			return nil, errors.New(responseHttp.HttpErrInternal, err)
		} else if !valid {
			return nil, errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid MCC code"))
		}
	}

	// Return early: if MCC doesn't match industry combination
	if request.ParentIndustry != "" && request.ChildIndustry != "" && request.MCC != "" {
		if err := s.industrySvc.ValidateIndustryMCCCombination(ctx, request.ParentIndustry, request.ChildIndustry, request.MCC); err != nil {
			return nil, errors.New(responseHttp.HttpErrUnprocessableContent, err)
		}
	}

	// Return early: if Digital Status is invalid
	if request.DigitalStatus != "" && !industry.IsValidDigitalStatus(request.DigitalStatus) {
		return nil, errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("digital status must be 'Digital' or 'Non-digital'"))
	}

	// Return early: if Country of Entity is invalid
	if request.CountryOfEntity != "" && !industry.IsValidCountryEntity(request.CountryOfEntity) {
		return nil, errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid country code for country of entity"))
	}

	err = merchant.UpdateMerchant(request)
	if err != nil {
		s.logger.Error(ctx, "error update submerchant", logger.Error(err), logger.Any("request", request))
		return nil, errors.New(responseHttp.HttpErrUnprocessableContent, constant.ErrUpdateMerchant)
	}
	if err = s.repo.Update(ctx, merchant); err != nil {
		s.logger.Error(ctx, "error update submerchant", logger.Error(err), logger.Any("request", request))
		return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrUpdateMerchant)
	}
	if needSendLinkInvitation {
		_, err = s.UserSvc.CreateMerchantUser(ctx, &user.MerchantUserRequest{
			Email:        request.PICEmail,
			Name:         request.PICName,
			Role:         constant.RoleAdmin,
			MerchantId:   merchant.UUID,
			MerchantName: merchant.Name,
			Invitation:   slices.Contains([]string{constant.KYCStatusNotRequired, constant.KYCStatusApproved}, merchant.KYCStatus.String),
		})
		if err != nil {
			return nil, err
		}
	}
	return merchant, nil
}
