package transferService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *TransferService) GetList(ctx context.Context, req *transfer.GetTransferListRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/GetList")
	defer segment.End()

	err := req.ValidateAndAdjust()
	if err != nil {
		return nil, errPkg.New(response.HttpErrRequest, err)
	}
	data, total, err := s.repo.GetList(ctx, req, page, perPage)
	if err != nil {
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetTransferList)
	}
	respData, err := s.buildListResponse(ctx, req.MerchantID, data)
	if err != nil {
		return nil, err
	}

	meta := commonModel.NewMeta(page, perPage, total)
	return &commonModel.PaginationResponse{Data: respData, Meta: *meta}, nil
}

func (s *TransferService) buildListResponse(ctx context.Context, merchantId string, data []*transfer.Transfer) ([]*transfer.ListTransferResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/buildListResponse")
	defer segment.End()

	var (
		merchantIDs      = []string{}
		merchantsNameMap = map[string]string{}
		respData         = []*transfer.ListTransferResponse{}
	)
	merchantIDs = append(merchantIDs, merchantId)
	for _, v := range data {
		if v.MerchantID.String() != merchantId {
			merchantIDs = append(merchantIDs, v.MerchantID.String())
		}
		if v.RecipientID.String() != merchantId {
			merchantIDs = append(merchantIDs, v.RecipientID.String())
		}
	}
	listMerchants, err := s.merchantSvc.GetMerchantsByIDs(ctx, merchantIDs)
	if err != nil {
		s.logger.Error(ctx, "error when get merchants by ids", logger.Error(err), logger.Any("merchantIds", merchantIDs))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetTransferList)
	}

	for _, v := range listMerchants {
		merchantsNameMap[v.UUID] = v.Name
	}
	for _, v := range data {
		respData = append(respData, v.ToListTransferResponse(merchantId, merchantsNameMap))
	}

	return respData, nil
}
