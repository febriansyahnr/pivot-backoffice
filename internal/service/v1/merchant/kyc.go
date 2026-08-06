package merchant

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var kycStatusTransitions = map[string][]string{
	constant.KYCStatusWaitingForDocument: {constant.KYCStatusInReview},
	constant.KYCStatusInReview:           {constant.KYCStatusApproved, constant.KYCStatusRejected, constant.KYCStatusNeedResubmission},
	constant.KYCStatusNeedResubmission:   {constant.KYCStatusInReview},
}

// UpdateKYC updates the KYC (Know Your Customer) of a merchant.
// it will change the merchant information and merchant status based on the KYC status.
// when kyc was changed, then it should update the cache
// this function used for internal purpose
func (s *MerchantService) UpdateKYC(ctx context.Context, payload merchantModel.UpdateMerchantKYCRequest) (*merchantModel.UpdateMerchantKYCResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateKYC")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, payload.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error when find merchant by id", logger.Error(err), logger.Any("payload", payload))
		return nil, err

	} else if merchant == nil {
		s.logger.Error(ctx, "merchant not found", logger.Any("payload", payload))
		return nil, constant.ErrMerchantNotFound
	}

	payload.MerchantStatus = merchant.Status
	if !s.AllowedChangeKYCStatus(merchant.KYCStatus.String, payload.KYCStatus) {
		s.logger.Error(ctx, "invalid kyc status change", logger.Any("payload", payload))
		return nil, pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenChangeKYCStatus)
	}

	if payload.KYCStatus == constant.KYCStatusApproved {
		payload.MerchantStatus = constant.MerchantStatusActive

		mid, err := s.repo.GenerateNewMID(ctx)
		if err != nil {
			s.logger.Error(ctx, "error when generate new MID", logger.Error(err), logger.Any("payload", payload))
			return nil, err
		}
		payload.MID = &mid
	}

	if err = s.repo.UpdateKYC(ctx, payload); err != nil {
		s.logger.Error(ctx, "error when update kyc", logger.Error(err), logger.Any("payload", payload))
		return nil, err
	}

	subAccountRegCallback := &pb.SubAccountRegistration{
		SubAccountId:        merchant.UUID,
		SubAccountStatus:    payload.MerchantStatus,
		SubAccountKycStatus: payload.KYCStatus,
		UpdatedAt:           timestamppb.New(time.Now().UTC()),
	}

	eventStatus := constant.CallbackStatusPending
	switch payload.KYCStatus {
	case constant.KYCStatusRejected:
		eventStatus = constant.CallbackStatusRejected

	case constant.KYCStatusApproved:
		eventStatus = constant.CallbackStatusApproved
	}
	callback := &pb.ProcessCallbackRequest{
		Name:       constant.CallbackNameSubAccountRegistration,
		Event:      fmt.Sprintf(constant.CallbackEventSubAccountRegistrationPattern, eventStatus),
		MerchantId: merchant.ParentID.String,
	}
	callback.Request, _ = anypb.New(subAccountRegCallback)

	_ = s.rabbitMqExt.PublishMerchantCallback(ctx, callback)

	if payload.KYCStatus == constant.KYCStatusApproved {
		picUser, _ := s.UserSvc.FindUserByEmail(ctx, merchant.PICEmail)
		if picUser != nil && picUser.MerchantId == merchant.UUID {
			ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, merchant.UUID)

			err = s.UserSvc.SendGeneratedInvitationURL(ctx, &user.SendGeneratedInvitationRequest{
				Inviter:      merchant.Name,
				Email:        merchant.PICEmail,
				MerchantName: merchant.Name,
				MerchantID:   merchant.UUID,
				UserID:       picUser.UUID,
				UserName:     picUser.Name,
			})
			if err != nil {
				s.logger.Error(ctx, "Failed while send generated invitation URL", logger.Error(err))
			}
		}
	}
	_ = s.redis.Del(ctx, fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchant.UUID)).Err()

	s.logger.Info(ctx, "KYC updated", logger.Any("payload", payload))

	return &merchantModel.UpdateMerchantKYCResponse{
		UUID:   payload.MerchantID,
		Status: payload.KYCStatus,
	}, nil
}

// AllowedChangeKYCStatus checks if the transition from the current KYC status
// to the desired KYC status is allowed based on predefined rules.
func (s *MerchantService) AllowedChangeKYCStatus(before, after string) bool {
	dest, exist := kycStatusTransitions[before]
	if !exist {
		return false
	}

	return slices.Contains(dest, after)
}
