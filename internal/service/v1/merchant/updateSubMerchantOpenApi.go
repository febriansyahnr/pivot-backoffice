package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) UpdateSubMerchantOpenApi(ctx context.Context, request *merchantModel.UpdateMerchantOpenApiRequest) (*merchantModel.SubMerchantResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateSubMerchantOpenApi")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, request.ID)
	if err != nil {
		s.logger.Error(ctx, "error when finding merchant by id", logger.Error(err), logger.Any("request", request))
		return nil, errors.New(responseHttp.HttpErrDatabase, constant.ErrUpdateMerchant)

	} else if merchant == nil || merchant.ParentID.String != request.ParentId {
		// Treat parent mismatch as "not found" to avoid leaking existence of sibling parents'
		// sub-merchants (consistent with Detail endpoint and GitHub's private-repo pattern).
		return nil, errors.New(responseHttp.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	merchant, err = s.UpdateSubMerchant(ctx, &merchantModel.UpdateMerchantRequest{
		ID:            request.ID,
		Name:          request.Name,
		Description:   request.Description,
		Address:       request.Address,
		PostCode:      request.PostCode,
		Logo:          request.Logo,
		MerchantEmail: request.MerchantEmail,
		MerchantPhone: request.MerchantPhone,
	})
	if err != nil {
		s.logger.Error(ctx, "error updating submerchant", logger.Error(err), logger.Any("request", request))
		return nil, err
	}
	return merchant.ToSubMerchantResponse(), nil
}
