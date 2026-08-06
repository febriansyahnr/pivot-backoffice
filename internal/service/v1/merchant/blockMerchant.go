package merchant

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) BlockMerchant(ctx context.Context, merchantId string) (*merchantModel.BlockMerchantResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/BlockMerchant")
	defer segment.End()

	var resp = &merchantModel.BlockMerchantResponse{
		BlockedMerchantDetails: merchantModel.BlockedMerchantDetails{},
		SubMerchants:           make([]merchantModel.BlockedMerchantDetails, 0),
	}

	merchant, err := s.FindMerchantByID(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrIncorrectMerchantID)
	}

	merchantsToBlock := []*merchantModel.Merchant{merchant}

	// if merchant parent id not valid, then merchant have sub merchant
	if !merchant.ParentID.Valid {
		subMerchants, err := s.repo.GetSubMerchantsByParentID(ctx, merchantId)
		if err != nil {
			return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrDatabaseGetData)
		}

		merchantsToBlock = append(merchantsToBlock, subMerchants...)
	}

	for idx, subMerchant := range merchantsToBlock {
		subMerchant.SetBlocked()
		updatedMerchant := merchantModel.BuildUpdateMerchantRequest(subMerchant)

		_, err = s.UpdateSubMerchant(ctx, &updatedMerchant)
		if err != nil {
			s.logger.Error(ctx, "BlockMerchant - error when updating sub merchant", logger.String("merchantId", subMerchant.UUID), logger.Error(err))

			// should skip if we failed to block the merchant
			continue
		}

		// after status change, need to delete merchant cache status
		err = s.redis.Del(ctx, fmt.Sprintf(constant.MerchantStatusByIDCacheKey, subMerchant.UUID)).Err()
		if err != nil {
			s.logger.Error(ctx, "BlockMerchant - failed to delete merchant status cache", logger.String("merchantId", subMerchant.UUID), logger.Error(err))

			// should skip too, if failed to delete the cache
			continue
		}

		// block the static VA from the merchant
		blockVAResponse, err := s.snapCoreRepo.BlockVirtualAccount(ctx, &snapCoreModel.BlockVirtualAccountRequest{
			MerchantID: subMerchant.UUID,
		})
		if err != nil {
			s.logger.Error(ctx, "BlockMerchant - error when block static VA sub merchant", logger.String("merchantId", subMerchant.UUID), logger.Error(err))
		}

		s.logger.Info(ctx, "BlockMerchant - block VA response from snap core", logger.String("merchantId", subMerchant.UUID), logger.Any("response", blockVAResponse))

		// idx == 0 means that it is parent merchant
		if idx == 0 {
			resp.BlockedMerchantDetails = merchantModel.BlockedMerchantDetails{
				MerchantId:   merchant.UUID,
				MerchantName: merchant.Name,
				BlockedAt:    time.Now().UTC(),
			}

			continue
		}

		resp.SubMerchants = append(resp.SubMerchants, merchantModel.BlockedMerchantDetails{
			MerchantId:   subMerchant.UUID,
			MerchantName: subMerchant.Name,
			BlockedAt:    time.Now().UTC(),
		})
	}

	return resp, nil
}

func (s *MerchantService) UnblockMerchant(ctx context.Context, merchantId string) (*merchantModel.UnblockMerchantResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UnblockMerchant")
	defer segment.End()

	var resp *merchantModel.UnblockMerchantResponse
	merchant, err := s.FindMerchantByID(ctx, merchantId)
	if err != nil {
		return nil, err
	}
	if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrIncorrectMerchantID)
	}

	merchantsToUnblock := []*merchantModel.Merchant{merchant}
	// if merchant parent id not valid, then merchant have sub merchant
	if !merchant.ParentID.Valid {
		subMerchants, err := s.repo.GetSubMerchantsByParentID(ctx, merchantId)
		if err != nil {
			s.logger.Error(ctx, "error when retrieve submerchants", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrDatabaseGetData)
		}
		merchantsToUnblock = append(merchantsToUnblock, subMerchants...)
	}

	for idx, merchant := range merchantsToUnblock {
		merchant.Unblock()
		err = s.repo.Update(ctx, merchant)
		if err != nil {
			s.logger.Error(ctx, "UnblockMerchant - error when updating sub merchant", logger.String("merchantId", merchant.UUID), logger.Error(err))
			continue
		}

		err = s.redis.Del(ctx, fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchant.UUID)).Err()
		if err != nil {
			s.logger.Error(ctx, "UnblockMerchant - failed to delete merchant status cache", logger.String("merchantId", merchant.UUID), logger.Error(err))
			continue
		}

		totalActiveTopUpReferences, err := s.merchantTopUpRepo.CountActiveMerchantTopUpReferences(ctx, &merchantTopUp.GetMerchantTopUpReferencesRequest{
			MerchantID: merchant.UUID,
		})
		if err != nil {
			s.logger.Error(ctx, "UnblockMerchant - error when getting merchant top up references", logger.String("merchantId", merchant.UUID), logger.Error(err))
		}
		if totalActiveTopUpReferences > 0 {
			unblockVAResponse, err := s.snapCoreRepo.UnblockVirtualAccount(ctx, &snapCoreModel.UnblockVirtualAccountRequest{
				MerchantID: merchant.UUID,
			})
			if err != nil {
				s.logger.Error(ctx, "UnblockMerchant - error when unblock static VA sub merchant", logger.String("merchantId", merchant.UUID), logger.Error(err))
			}
			s.logger.Info(ctx, "UnblockMerchant - unblock VA response from snap core", logger.String("merchantId", merchant.UUID), logger.Any("response", unblockVAResponse))
		}

		if resp == nil {
			resp = &merchantModel.UnblockMerchantResponse{
				SubMerchants: make([]merchantModel.UnblockedMerchantDetails, 0),
			}
		}
		if idx == 0 {
			resp.UnblockedMerchantDetails = merchantModel.UnblockedMerchantDetails{
				MerchantId:   merchant.UUID,
				MerchantName: merchant.Name,
				UnblockedAt:  time.Now().UTC(),
			}
			continue
		}

		resp.SubMerchants = append(resp.SubMerchants, merchantModel.UnblockedMerchantDetails{
			MerchantId:   merchant.UUID,
			MerchantName: merchant.Name,
			UnblockedAt:  time.Now().UTC(),
		})
	}

	return resp, nil
}
