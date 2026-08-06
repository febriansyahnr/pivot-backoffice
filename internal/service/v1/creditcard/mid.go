package creditcard

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func (s *CreditCardService) GetMIDList(ctx context.Context, request *creditcardModel.GetMIDListRequest) (*commonModel.PaginationResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/GetMIDList")
	defer span.End()

	data, err := s.creditcardCoreProcessorRepo.GetMIDList(ctx, &creditcardCoreProcessorModel.GetMIDListRequest{
		Page:            request.Page,
		Limit:           request.PerPage,
		Mid:             request.Mid,
		Acquirer:        request.Acquirer,
		Name:            request.Name,
		Type:            request.Type,
		TransactionType: request.TransactionType,
		InstallmentType: request.InstallmentType,
		IsDefault:       request.IsDefault,
		IsActive:        request.IsActive,
	})
	if err != nil {
		return nil, err
	}

	return &commonModel.PaginationResponse{
		Data: util.MapSnakeToCamel(data.Results),
		Meta: commonModel.Meta{
			Page:       int64(data.Pagination.PageNumber),
			PerPage:    int64(data.Pagination.PageLimit),
			TotalPages: int64(data.Pagination.TotalPage),
			TotalItems: int64(data.Pagination.TotalRecord),
		},
	}, nil
}

func (s *CreditCardService) GetMIDMapList(ctx context.Context, limit, page int, merchantId string) (*commonModel.PaginationResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/GetMIDMapList")
	defer span.End()

	data, err := s.creditcardCoreProcessorRepo.GetMIDMapList(ctx, &creditcardCoreProcessorModel.GetMIDMapListRequest{
		Page:       page,
		Limit:      limit,
		MerchantId: merchantId,
	})
	if err != nil {
		return nil, err
	}

	return &commonModel.PaginationResponse{
		Data: util.MapSnakeToCamel(data.Results),
		Meta: commonModel.Meta{
			Page:       int64(data.Pagination.PageNumber),
			PerPage:    int64(data.Pagination.PageLimit),
			TotalPages: int64(data.Pagination.TotalPage),
			TotalItems: int64(data.Pagination.TotalRecord),
		},
	}, nil
}

func (s *CreditCardService) CreateMID(ctx context.Context, request *creditcardModel.CreateMIDRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/CreateMID")
	defer span.End()

	_, err := s.creditcardCoreProcessorRepo.CreateMID(ctx, request.ToCreditCardCoreRequest())
	if err != nil {
		return err
	}

	return nil
}

func (s *CreditCardService) UpdateMID(ctx context.Context, request *creditcardModel.UpdateMIDRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/UpdateMID")
	defer span.End()

	// TODO:
	// Validate if there is installment plans related to this mid
	_, err := s.creditcardCoreProcessorRepo.UpdateMID(ctx, request.ToCreditCardCoreRequest())
	if err != nil {
		return err
	}

	return nil
}

func (s *CreditCardService) GetMIDDetail(ctx context.Context, midId string) (*creditcardCoreProcessorModel.MIDResponseData, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/GetMIDDetail")
	defer span.End()

	data, err := s.creditcardCoreProcessorRepo.GetMID(ctx, midId)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *CreditCardService) ValidateMIDInstallmentBins(ctx context.Context, request *creditcardModel.ValidateMIDInstallmentBinsRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/ValidateMIDInstallmentBins")
	defer span.End()

	err := s.creditcardCoreProcessorRepo.ValidateMidInstallmentBins(ctx, &creditcardCoreProcessorModel.ValidateMIDInstallmentBinsRequest{
		MidID: request.MidID,
		Bins:  request.Bins,
	})
	if err != nil {
		return err
	}

	return nil
}
