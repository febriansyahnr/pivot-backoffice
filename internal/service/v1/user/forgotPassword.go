package user

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) ForgotPassword(ctx context.Context, email string) (token string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ForgotPassword")
	defer segment.End()

	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if user == nil {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("email not registered"))

	} else if !user.Blocked.Time.IsZero() || user.Status == constant.UserStatusBlocked {
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("user has been blocked"))

	} else if user.Status == constant.UserStatusInvited {
		return "", pkgErrs.New(response.HttpStatusErrorConflict, constant.ErrUserInvitedStatus)

	} else if user.MerchantStatus.String == constant.MerchantStatusBlocked {
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("merchant is blocked"))

	} else if user.MerchantStatus.String == constant.MerchantStatusInactive {
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("merchant is inactive"))

	} else if user.MerchantStatus.String == constant.MerchantStatusDeactivated {
		return "", pkgErrs.New(response.HttpErrForbidden, errors.New("merchant is deactivated"))

	} else if user.MerchantStatus.String == constant.MerchantStatusClosed {
		return "", pkgErrs.New(response.HttpErrUnauthorized, errors.New("Merchant status is closed. Reason: "+user.ReasonStatus.String))
	}

	ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, user.MerchantId)

	token, err = s.otpSvc.GenerateOTPCode(ctx, user.UUID, user.Email, constant.OTPIdentifierForgotPassword)
	if err != nil {
		s.logger.Error(ctx, "Forgot Password: generate OTP code", logger.Error(err))
	}
	return
}

func (s *UserService) ResetPassword(ctx context.Context, id, password string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ResetPassword")
	defer segment.End()

	_, err = s.userRepo.ChangePassword(ctx, id, encryption.EncryptPassword(password))
	return
}
