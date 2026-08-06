package tnc

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *TNCService) CreateTNCVersion(
	ctx context.Context,
	req *tncModel.CreateTNCVersionRequest,
) (*tncModel.TNCVersionResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/CreateTNCVersion")
	defer span.End()

	existing, err := s.repo.GetTNCVersionByVersion(ctx, req.Version)
	if err != nil {
		s.logger.Error(ctx, "error when checking existing tnc version", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if existing != nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrTNCVersionAlreadyExists)
	}

	version := req.ToTNCVersion()
	if err := s.repo.CreateTNCVersion(ctx, version); err != nil {
		if strings.Contains(err.Error(), "1062") && strings.Contains(err.Error(), "Duplicate entry") {
			return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrTNCVersionAlreadyExists)
		}
		s.logger.Error(ctx, "error when creating tnc version", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return version.ToResponse(), nil
}

func (s *TNCService) ActivateTNCVersion(ctx context.Context, id string) (*tncModel.TNCVersionResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/ActivateTNCVersion")
	defer span.End()

	return s.setTNCVersionActive(ctx, id, true)
}

func (s *TNCService) DeactivateTNCVersion(ctx context.Context, id string) (*tncModel.TNCVersionResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/DeactivateTNCVersion")
	defer span.End()

	return s.setTNCVersionActive(ctx, id, false)
}

func (s *TNCService) setTNCVersionActive(ctx context.Context, id string, active bool) (*tncModel.TNCVersionResponse, error) {
	version, err := s.repo.GetTNCVersionByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when getting tnc version for activation", logger.Error(err), logger.String("id", id))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if version == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrTNCVersionNotFound)
	}

	if active {
		if err := s.repo.DeactivateAllTNCVersions(ctx); err != nil {
			s.logger.Error(ctx, "error when deactivating other tnc versions", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, err)
		}
	}

	version.IsActive = active
	if err := s.repo.UpdateTNCVersion(ctx, version); err != nil {
		s.logger.Error(ctx, "error when setting tnc version active flag", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return version.ToResponse(), nil
}

func (s *TNCService) ListTNCVersions(
	ctx context.Context,
	q *tncModel.TNCVersionQuery,
) (*commonModel.PaginationResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/ListTNCVersions")
	defer span.End()

	list, total, err := s.repo.ListTNCVersions(ctx, q)
	if err != nil {
		s.logger.Error(ctx, "error when listing tnc versions", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrGetTNCVersionList)
	}

	responses := make([]*tncModel.TNCVersionResponse, 0, len(list))
	for _, v := range list {
		responses = append(responses, v.ToResponse())
	}

	return &commonModel.PaginationResponse{
		Data: responses,
		Meta: *commonModel.NewMeta(q.Page, q.PageSize, int64(total)),
	}, nil
}

func (s *TNCService) GetTNCVersion(ctx context.Context, id string) (*tncModel.TNCVersionResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/GetTNCVersion")
	defer span.End()

	version, err := s.repo.GetTNCVersionByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when getting tnc version", logger.Error(err), logger.String("id", id))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if version == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrTNCVersionNotFound)
	}

	return version.ToResponse(), nil
}
