package role

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *RoleService) Delete(ctx context.Context, merchantID, roleID string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/Delete")
	defer segment.End()

	if res, err := s.repo.FindRoleByID(ctx, roleID); err != nil {
		s.logger.Error(ctx, "error when finding role by id", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)

	} else if res == nil || res.MerchantID.String != merchantID {
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrRoleNotFound)
	}

	if total, err := s.userRoleRepo.TotalActiveUsersByRoleID(ctx, roleID); err != nil {
		s.logger.Error(ctx, "error when get total active users by role id", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)

	} else if total > 0 {
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrRoleCannotBeDeleted)
	}

	if ctx, err = s.repo.BeginTransaction(ctx); err != nil {
		s.logger.Error(ctx, "error when execute begin transaction", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.repo.RollbackTransaction(ctx); e != nil {
				err = pkgErrs.New(response.HttpErrDatabase, e)
				s.logger.Error(ctx, "error when execute rollback transaction", logger.Error(err))
			}
		}
	}()

	if err = s.roleMenuPermRepo.Delete(ctx, roleID); err != nil {
		s.logger.Error(ctx, "error when delete role menu permission", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	if err = s.repo.Delete(ctx, roleID); err != nil {
		s.logger.Error(ctx, "error when delete role", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	if err = s.repo.CommitTransaction(ctx); err != nil {
		s.logger.Error(ctx, "error when execute commit transaction", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	return
}
