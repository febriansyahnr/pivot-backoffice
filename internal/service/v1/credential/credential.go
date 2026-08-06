package credential

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	credModel "github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) Get(ctx context.Context, request *credModel.CredentialDashboardReq) (*credModel.CredentialDashboardResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/credential/Get")
	defer segment.End()

	resp, err := s.repo.Get(ctx, request.MerchantID)
	if err != nil {
		s.log.Error(ctx, "failed to get merchant data", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if resp == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}
	for i := 0; i < len(resp.ClientSecrets); i++ {
		resp.ClientSecrets[i].KeyName = fmt.Sprintf("Client Secret %d", i+1)
	}

	s.ActivityLog(ctx, &request.MerchantID, &request.UserID, request.Info, constant.ActivityUserAccessCredDashboard, nil)

	return resp.ToResponse(), nil
}
