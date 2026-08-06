package vendor

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *VendorService) Delete(ctx context.Context, uuid string) error {
	ctx, span := tracer.Start(ctx, "internal/service/v1/vendor/Delete")
	defer span.End()

	vendor, err := s.vendorRepo.GetByID(ctx, uuid)
	if err != nil {
		s.logger.Error(ctx, "error when getting vendor for delete", logger.Error(err), logger.Any("uuid", uuid))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	if vendor == nil {
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrVendorNotFound)
	}

	err = s.vendorRepo.Delete(ctx, uuid)
	if err != nil {
		s.logger.Error(ctx, "error when deleting vendor", logger.Error(err), logger.Any("uuid", uuid))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	return nil
}
