package vendor

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *VendorService) Update(ctx context.Context, request *vendorModel.UpdateVendorRequest) (*vendorModel.VendorResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/vendor/Update")
	defer span.End()

	vendor, err := s.vendorRepo.GetByID(ctx, request.UUID)
	if err != nil {
		s.logger.Error(ctx, "error when getting vendor for update", logger.Error(err), logger.Any("uuid", request.UUID))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if vendor == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrVendorNotFound)
	}

	vendor.Update(request)

	err = s.vendorRepo.Update(ctx, vendor)
	if err != nil {
		s.logger.Error(ctx, "error when updating vendor", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return vendor.ToResponse(), nil
}
