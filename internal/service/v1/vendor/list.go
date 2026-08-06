package vendor

import (
	"context"
	"math"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *VendorService) List(ctx context.Context, req *vendorModel.VendorQuery) (*commonModel.PaginationResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/vendor/List")
	defer span.End()

	vendors, total, err := s.vendorRepo.List(ctx, req)
	if err != nil {
		s.logger.Error(ctx, "error when listing vendors", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	list := make([]*vendorModel.VendorResponse, len(vendors))
	for i, v := range vendors {
		list[i] = v.ToResponse()
	}

	totalPage := int64(math.Ceil(float64(total) / float64(req.PageSize)))

	return &commonModel.PaginationResponse{
		Data: list,
		Meta: commonModel.Meta{
			Page:       req.Page,
			PerPage:    req.PageSize,
			TotalItems: int64(total),
			TotalPages: totalPage,
		},
	}, nil
}
