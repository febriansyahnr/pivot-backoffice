package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	response "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *MerchantService) ValidateSubMerchantParent(ctx context.Context, parentMerchantID, merchantID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/ValidateSubMerchantParent")
	defer segment.End()

	subMerchant, err := s.repo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedValidateSubMerchantParent)
	}

	if subMerchant == nil {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrSubMerchantNotFound)
	}

	if subMerchant.ParentID.String != parentMerchantID {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrIncorrectSubMerchantParent)
	}

	return nil
}
