package vendor

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *VendorService) Detail(ctx context.Context, uuid string) (*vendorModel.Vendor, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/vendor/Detail")
	defer span.End()

	vendor, err := s.vendorRepo.GetByID(ctx, uuid)
	if err != nil {
		s.logger.Error(ctx, "error when getting vendor detail", logger.Error(err), logger.Any("uuid", uuid))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if vendor == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrVendorNotFound)
	}

	return vendor, nil
}
