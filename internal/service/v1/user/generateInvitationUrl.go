package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commServicePb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/commService"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) SendGeneratedInvitationURL(ctx context.Context, request *userModel.SendGeneratedInvitationRequest) error {
	feature := constant.UserIdentifierUserInvitation
	dataKey := redisTokenKey(request.Email, feature, ":data")

	totalResend := 0
	if err := s.redis.HGet(ctx, dataKey, constant.UserInvitationTotalResendField).Scan(&totalResend); err != nil && !errors.Is(err, redisExt.ErrNil) {
		return pkgErrors.New(response.HttpErrDatabase, err)

	} else if (totalResend + 1) > feature.MexSendUserInvitation() {
		return pkgErrors.New(response.HttpErrTooManyRequest, errors.New("your request has exceeded the limit"))
	}

	// Rate Limiting
	limit := &redisExt.Limit{
		Rate:   1,
		Burst:  1,
		Period: 3 * time.Minute,
	}
	if resp, err := s.limiter.Allow(ctx, redisTokenKey(request.Email, feature), limit); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)

	} else if resp.Allowed == 0 {
		return pkgErrors.New(response.HttpErrResourceLocked, constant.ErrResourceLocked)
	}

	lastToken := ""
	if err := s.redis.HGet(ctx, dataKey, constant.UserInvitationLastTokenField).Scan(&lastToken); err != nil && !errors.Is(err, redisExt.ErrNil) {
		return pkgErrors.New(response.HttpErrDatabase, err)

	}

	if request.IsResend {
		totalResend += 1
		s.removePreviousInvitationToken(ctx, lastToken)
	}

	// Generate invitation url
	randomString, _ := util.GenerateUniqueRandomToken(32)
	token := encryption.GenerateHMAC(string(feature), request.Email) + "." + randomString
	invitationURL, err := s.generateInvitationURL(ctx, request, token)
	if err != nil {
		return err
	}

	userInvitationData := map[string]interface{}{
		constant.UserInvitationTotalResendField: totalResend,
		constant.UserInvitationLastTokenField:   token,
	}
	if errSet := s.redis.HSet(ctx, dataKey, userInvitationData).Err(); errSet != nil {
		return pkgErrors.New(response.HttpErrDatabase, errSet)
	}
	if errExpire := s.redis.Expire(ctx, dataKey, feature.ExpireDuration()).Err(); errExpire != nil {
		return pkgErrors.New(response.HttpErrDatabase, errExpire)
	}

	content, _ := structpb.NewStruct(map[string]any{
		"Inviter":           request.Inviter,
		"InvitationURL":     invitationURL,
		"DashboardGuideURL": s.config.MerchantPortalConfig.DashboardGuideURL,
		"LogoURL":           s.config.PaperCommunication.EmailLogoURL,
	})
	emailRequest := &commServicePb.EmailRequest{
		Event:                feature.Event(),
		From:                 feature.EmailSender(),
		To:                   request.Email,
		Subject:              fmt.Sprintf("You're invited to join %s team on %s", request.MerchantName, s.config.PaperCommunication.PlatformName),
		Content:              content,
		Priority:             commServicePb.EmailPriority_L0,
		ToBeRetriedOnFailure: true,
	}
	payload, _ := proto.Marshal(emailRequest)

	return s.rabbitMqExt.Publish(ctx, rabbitMqExt.CommServiceEmailRoutingKey, nil, payload)
}

func (s *UserService) generateInvitationURL(ctx context.Context, request *userModel.SendGeneratedInvitationRequest, token string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/GenerateInvitationURL")
	defer segment.End()

	feature := constant.UserIdentifierUserInvitation
	lockKey := redisTokenKey(request.Email, feature, ":lock")

	if can, err := s.redis.SetNX(ctx, lockKey, true, 10*time.Second).Result(); err != nil {
		return "", pkgErrors.New(response.HttpErrDatabase, err)
	} else if !can {
		return "", pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("the same request is in progress"))
	}
	defer func() {
		_ = s.redis.Del(ctx, lockKey)
	}()

	invitationData := &userModel.ValidateInvitationResponse{
		UserID:       request.UserID,
		UserName:     request.UserName,
		Email:        request.Email,
		MerchantName: request.MerchantName,
		MerchantID:   request.MerchantID,
	}

	_ = s.redis.Set(ctx, redisTokenKey("", feature, ":", constant.UserTokenNamespace, ":", token), invitationData, feature.ExpireDuration())

	merchantPortalInvitationURL := s.config.MerchantPortalConfig.UserInvitationURL + fmt.Sprintf("?token=%s", token)
	return merchantPortalInvitationURL, nil
}

func (s *UserService) removePreviousInvitationToken(ctx context.Context, token string) {
	key := redisTokenKey("", constant.UserIdentifierUserInvitation, ":", constant.UserTokenNamespace, ":", token)
	if errDel := s.redis.Del(ctx, key).Err(); errDel != nil {
		s.logger.Error(ctx, fmt.Sprintf("Error deleting key:%s", key), logger.Error(errDel))
	}
}
