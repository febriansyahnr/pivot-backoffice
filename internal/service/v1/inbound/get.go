package inboundService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *InboundService) GetList(ctx context.Context, filter *inboundModel.GetInboundFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/inbound/GetList")
	defer span.End()

	list, err := s.inboundRepo.GetList(ctx, filter)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	return list, nil
}

func (s *InboundService) GetByID(ctx context.Context, id string) (*inboundModel.InboundResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/inbound/GetByID")
	defer span.End()

	detail, err := s.inboundRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	return detail, nil
}

func (s *InboundService) GetSnapVersionByID(ctx context.Context, id string) (*inboundModel.InboundSnapVersionResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/inbound/GetSnapVersionByID")
	defer span.End()

	detail, err := s.inboundRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	if detail == nil {
		return nil, pkgErr.New(response.HttpErrNotFound, fmt.Errorf("inbound detail not found"))
	}

	if !detail.SnapCompatibility {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrRequestIsNotCompatibleWithSnapVersion)
	}

	return detail.ToSnapVersionResponse(), nil
}
