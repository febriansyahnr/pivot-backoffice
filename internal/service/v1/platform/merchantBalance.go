package platformService

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PlatformService) GetSubMerchantBalances(ctx context.Context, request *platform.GetBulkBalanceRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "/internal/service/v1/platform/GetSubMerchantBalances")
	defer segment.End()

	subMerchantListResponse, err := s.merchantService.ListSubMerchantByParentID(ctx, &merchant.SubMerchantListFilter{
		ParentId: request.MerchantID,
	}, int64(request.Page), int64(request.PerPage))
	if err != nil {
		s.logger.Error(ctx, "error when retrieve submerchant by parent ids", logger.Error(err), logger.Any("request", request))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetSubMerchantList)
	}

	subMerchantList, ok := subMerchantListResponse.Data.([]*merchant.Merchant)
	if !ok {
		s.logger.Error(ctx, "invalid response data compared to what anticipated", logger.Any("response", subMerchantListResponse))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetSubMerchantList)
	}
	merchantIds := make([]uuid.UUID, len(subMerchantList))
	for i, subMerchant := range subMerchantList {
		merchantIds[i] = uuid.MustParse(subMerchant.UUID)
	}

	merchantsBalanceResponseMap, err := s.orchestratorService.GetMerchantBulkBalances(ctx, &account_model.GetBulkBalanceRequest{
		MerchantIDs: merchantIds,
		Usecase:     request.Usecase,
	})
	if err != nil {
		s.logger.Error(ctx, "error when retrieve merchant bulk balance", logger.Error(err), logger.Any("request", request))
		return nil, err
	}

	responseData := make([]platform.MerchantBalanceResponse, len(subMerchantList))
	for i, subMerchant := range subMerchantList {
		balance := merchantsBalanceResponseMap[subMerchant.UUID]
		balanceResponse := platform.MerchantBalanceResponse{
			MerchantID: subMerchant.UUID,
			AvailableBalance: &platform.PlatformAvailableBalanceResponse{
				Value:    0,
				Currency: constant.CurrencyIDR,
			},
		}
		if balance != nil {
			balanceResponse.AvailableBalance.Value = balance.Balance
		}

		responseData[i] = balanceResponse
	}

	return &commonModel.PaginationResponse{
		Data: responseData,
		Meta: subMerchantListResponse.Meta,
	}, nil
}
