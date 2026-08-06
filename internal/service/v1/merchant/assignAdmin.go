package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
)

func (s *MerchantService) AssignSubMerchantAdmin(ctx context.Context, payload *merchant.SubMerchantAdminRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/AssignSubMerchantAdmin")
	defer segment.End()

	merchant, _ := s.repo.FindMerchantByID(ctx, payload.MerchantId)

	merchantName := ""
	if merchant != nil {
		merchantName = merchant.Name
	}

	createUserRequest := &userModel.MerchantUserRequest{
		Email:        payload.Email,
		Name:         payload.Name,
		Role:         constant.RoleAdmin,
		MerchantId:   payload.MerchantId,
		MerchantName: merchantName,
		Invitation:   payload.Invitation,
	}
	_, err := s.UserSvc.CreateMerchantUser(ctx, createUserRequest)
	return err
}
