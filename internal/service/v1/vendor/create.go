package vendor

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *VendorService) Create(ctx context.Context, request *vendorModel.CreateVendorRequest) (*vendorModel.VendorResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/vendor/Create")
	defer span.End()

	vendor := request.ToVendor()

	err := s.vendorRepo.Create(ctx, vendor)
	if err != nil {
		if strings.Contains(err.Error(), "1062") && strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "vendors_merchant_name_uniq_comp_idx") {
			return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrVendorNameAlreadyExists)
		}
		s.logger.Error(ctx, "error when creating vendor", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return vendor.ToResponse(), nil
}
