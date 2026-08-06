package merchant

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
)

func (s *MerchantService) BulkCreateSubMerchant(ctx context.Context, request *merchantModel.BulkCreateSubMerchantRequest) (*merchantModel.BulkCreateSubMerchantResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/BulkCreateSubMerchant")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, request.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "error when get merchant by id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	var (
		batchSize   = 10 //15 // Batch Size
		totalFailed int
		results     []merchantModel.BulkCreateSubMerchantDetailResponse
		sessionId   = uuid.NewString()
	)
	totalBatch := len(request.SubmerchantDetails)/batchSize + 1
	for i := 0; i < totalBatch; i++ {
		processBulkRequest := merchantModel.ProcessBulkCreateSubMerchantRequest{
			ID:                 sessionId,
			MerchantId:         request.MerchantId,
			KYCType:            request.KYCType,
			SubmerchantDetails: []merchantModel.BulkSubMerchantDetailRequest{},
			Batch:              i,
			FileName:           request.FileName,
		}
		for row := i * batchSize; row < (i*batchSize)+batchSize; row++ {
			if row > len(request.SubmerchantDetails)-1 {
				break
			}
			submerchantDetails := request.SubmerchantDetails[row]
			merchantRequest := &merchantModel.MerchantRequest{
				Name:              submerchantDetails[0],
				ShortName:         submerchantDetails[1],
				Logo:              submerchantDetails[2],
				MerchantEmail:     submerchantDetails[3],
				MerchantPhone:     submerchantDetails[4],
				BusinessCountry:   submerchantDetails[5],
				BusinessType:      submerchantDetails[6],
				BusinessStructure: submerchantDetails[7],
				PICName:           submerchantDetails[8],
				PICPhone:          submerchantDetails[9],
				PICEmail:          submerchantDetails[10],
				Address:           submerchantDetails[11],
				PostCode:          submerchantDetails[12],
				BankAccount: &merchantModel.MerchantBankAccountRequest{
					AccountNumber: submerchantDetails[13],
					ChannelCode:   submerchantDetails[14],
				},
				KYCStatus:      request.KYCType,
				PICInvitation:  request.IsInvitePIC,
				MerchantStatus: constant.MerchantStatusActive,
				ParentID:       request.MerchantId,
				RequesterID:    request.MerchantId,
			}
			err := s.validator.Struct(merchantRequest)
			if err != nil {
				totalFailed++
				results = append(results, merchantModel.BulkCreateSubMerchantDetailResponse{
					Row:   row,
					Error: err.Error(),
				})
				continue
			} else {
				switch request.KYCType {
				case constant.MerchantKYCTypeKYC:
					merchantRequest.KYCStatus = constant.KYCStatusApproved // Auto Approved from this migration
				case constant.MerchantKYCTypeNonKYC:
					merchantRequest.KYCStatus = constant.KYCStatusNotRequired
				}

				processBulkRequest.SubmerchantDetails = append(processBulkRequest.SubmerchantDetails, merchantModel.BulkSubMerchantDetailRequest{
					Row:    row,
					Detail: *merchantRequest,
				})
			}

			_ = s.storeBulkCreateSubMerchantResult(ctx, &processBulkRequest, results)
		}

		err = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SubMerchantBulkCreateRoutingKey, nil, processBulkRequest)
		if err != nil {
			s.logger.Error(ctx, "error when publish to rabbitmq", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, err)
		}
	}

	return &merchantModel.BulkCreateSubMerchantResponse{
		ID:          sessionId,
		TotalFailed: totalFailed,
		Results:     results,
	}, nil
}

func (s *MerchantService) ProcessBulkCreateSubMerchant(ctx context.Context, request *merchantModel.ProcessBulkCreateSubMerchantRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/ProcessBulkCreateSubMerchant")
	defer segment.End()

	var (
		results []merchantModel.BulkCreateSubMerchantDetailResponse
	)
	for _, submerchantDetail := range request.SubmerchantDetails {
		merchant, err := s.CreateSubMerchant(ctx, &submerchantDetail.Detail)
		if err != nil {
			s.logger.Error(ctx, "error when create submerchant", logger.Error(err), logger.String("sessiondId", request.ID), logger.Int("batch", request.Batch), logger.Int("row", submerchantDetail.Row))
			results = append(results, merchantModel.BulkCreateSubMerchantDetailResponse{
				Row:   submerchantDetail.Row,
				Error: err.Error(),
			})
			continue
		}
		results = append(results, merchantModel.BulkCreateSubMerchantDetailResponse{
			MerchantID:   merchant.UUID,
			MerchantName: merchant.Name,
			Row:          submerchantDetail.Row,
		})
	}

	return s.storeBulkCreateSubMerchantResult(ctx, request, results)
}

func (s *MerchantService) GetBulkCreateSubMerchantSummary(ctx context.Context, request *merchantModel.GetBulkCreateSubMerchantSummaryRequest) (*merchantModel.BulkCreateSubMerchantResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetBulkCreateSubMerchantSummary")
	defer segment.End()

	redisKey := fmt.Sprintf(constant.MerchantBulkCreateSubMerchantSessionIDCacheKey, request.MerchantId, request.ID)
	fileNameArr, err := s.redis.Client().ZRange(ctx, redisKey, 0, 1).Result()
	if err != nil {
		s.logger.Error(ctx, "error when retrieve file name from redis", logger.Error(err), logger.String("redisKey", redisKey))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}
	resultArr, err := s.redis.Client().ZRange(ctx, redisKey, 1, -1).Result()
	if err != nil {
		s.logger.Error(ctx, "error when retrieve result from redis", logger.Error(err), logger.String("redisKey", redisKey))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}
	var (
		results                   []merchantModel.BulkCreateSubMerchantDetailResponse
		totalFailed, totalSuccess int
	)
	for _, result := range resultArr {
		var detail merchantModel.BulkCreateSubMerchantDetailResponse
		err := json.Unmarshal([]byte(result), &detail)
		if err != nil {
			s.logger.Error(ctx, "error when unmarshal result from redis", logger.Error(err), logger.String("redisKey", redisKey))
			return nil, pkgErrs.New(response.HttpErrInternal, err)
		}
		results = append(results, detail)
		if detail.Error != "" {
			totalFailed++
		} else if detail.MerchantID != "" || detail.MerchantName != "" {
			totalSuccess++
		}
	}
	response := &merchantModel.BulkCreateSubMerchantResponse{
		ID:           request.ID,
		TotalFailed:  totalFailed,
		TotalSuccess: totalSuccess,
		Results:      results,
	}
	if len(fileNameArr) > 0 {
		response.FileName = fileNameArr[0]
	}
	return response, nil
}

func (s *MerchantService) storeBulkCreateSubMerchantResult(ctx context.Context, request *merchantModel.ProcessBulkCreateSubMerchantRequest, results []merchantModel.BulkCreateSubMerchantDetailResponse) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/storeBulkCreateSubMerchantResult")
	defer segment.End()

	if len(results) == 0 {
		return nil
	}

	redisKey := fmt.Sprintf(constant.MerchantBulkCreateSubMerchantSessionIDCacheKey, request.MerchantId, request.ID)
	for _, result := range results {
		b, _ := json.Marshal(result)
		_, err := s.redis.Client().ZAdd(ctx, redisKey, redis.Z{
			Score:  float64(result.Row),
			Member: string(b),
		}).Result()
		if err != nil {
			s.logger.Error(ctx, "error when store to redis", logger.Error(err), logger.Any("rowDetail", result))
		}
	}
	_, err := s.redis.Client().ZAdd(ctx, redisKey, redis.Z{
		Score:  -1,
		Member: request.FileName,
	}).Result()
	if err != nil {
		s.logger.Error(ctx, "error when store filename to redis", logger.Error(err))
	}

	_, err = s.redis.Expire(ctx, redisKey, constant.MerchantBulkCreateSubMerchantSessionIDCacheTTL).Result()
	if err != nil {
		s.logger.Error(ctx, "error when expire redis", logger.Error(err))
	}
	return nil
}
