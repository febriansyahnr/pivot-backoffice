package orchestrator_service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *OrchestratorService) GetPendingBalance(ctx context.Context, merchantID, balanceName string) (float64, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetPendingBalance")
	defer span.End()

	merchantUUID, err := uuid.Parse(merchantID)
	if err != nil {
		s.logger.Error(ctx, "error parsing merchant id", logger.Error(err))
		return 0, pkgErrors.New(httpResponse.HttpErrRequest, err)
	}

	disbursementAccount, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantUUID, balanceName)
	if err != nil {
		return 0, err
	}
	if disbursementAccount == nil {
		disbursementAccount, err = account_model.NewAccount(&account_model.NewAccountRequest{
			ReferenceID: merchantUUID,
			Usecase:     balanceName,
			Currency:    "IDR",
			UserType:    constant.UserTypeMerchant,
		})
		if err != nil {
			return 0, pkgErrors.New(httpResponse.HttpErrRequest, err)
		}

		if err = s.accountRepo.Create(ctx, disbursementAccount); err != nil {
			return 0, err
		}
	}

	aggregateTransactionRequest := &orchestrator_model.GetAggregateRequest{
		MerchantID: merchantUUID,
		AccountID:  disbursementAccount.UUID,
		Statuses:   []string{constant.StatusPending},
		StartAt:    nil,
		EndAt:      nil,
	}

	earliestUpdatedAt, err := s.GetCachedEarliestUpdatedAt(ctx, aggregateTransactionRequest)
	if err != nil {
		return 0, err
	}
	aggregateTransactionRequest.StartAt = &earliestUpdatedAt
	aggregateTransactionRequest.EndAt = util.ValueToPtr(time.Now())

	// Special handling for PAYMENT type
	if balanceName == constant.ProductPayment ||
		balanceName == constant.ReferenceWallet ||
		balanceName == constant.TypeVirtualTerminal {
		pendingBalance, err := s.accountTransactionRepo.CalculatePendingBalance(ctx, aggregateTransactionRequest)
		if err != nil {
			return 0, err
		}
		return pendingBalance, nil
	}

	aggregateTransaction, err := s.accountTransactionRepo.GetAggregateTransactions(ctx, aggregateTransactionRequest)
	if err != nil {
		return 0, err
	}

	pendingBalance := aggregateTransaction.SumOfCredit - aggregateTransaction.SumOfDebit
	return pendingBalance, nil
}

func (s *OrchestratorService) GetCachedEarliestUpdatedAt(ctx context.Context, request *orchestrator_model.GetAggregateRequest) (time.Time, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetCachedEarliestUpdatedAt")
	defer span.End()

	var earliestUpdatedAt time.Time

	// If redis is not configured, skip caching and go directly to database
	if s.redis == nil {
		return s.accountTransactionRepo.GetEarliestUpdatedAt(ctx, request)
	}

	requestHash, err := s.generateRequestHash(request)
	if err != nil {
		s.logger.Error(ctx, "error generating request hash for cache key", logger.Error(err))
		return earliestUpdatedAt, err
	}
	cacheKey := fmt.Sprintf(constant.EarliestUpdatedAtCacheKey, request.MerchantID.String(), requestHash)

	cachedValue, err := s.redis.Get(ctx, cacheKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "error getting cached earliest updated at", logger.Error(err))
	}

	if cachedValue != "" {
		if err := json.Unmarshal([]byte(cachedValue), &earliestUpdatedAt); err == nil {
			s.logger.Debug(ctx, "retrieved cached earliest updated at", logger.Time("earliest_updated_at", earliestUpdatedAt))
			return earliestUpdatedAt, nil
		}
		s.logger.Error(ctx, "error unmarshaling cached earliest updated at", logger.Error(err))
	}

	earliestUpdatedAt, err = s.accountTransactionRepo.GetEarliestUpdatedAt(ctx, request)
	if err != nil {
		return earliestUpdatedAt, err
	}

	cachedData, err := json.Marshal(earliestUpdatedAt)
	if err != nil {
		s.logger.Error(ctx, "error marshaling earliest updated at for cache", logger.Error(err))
		return earliestUpdatedAt, nil
	}

	expiration := 24 * time.Hour
	if err := s.redis.Set(ctx, cacheKey, string(cachedData), expiration).Err(); err != nil {
		s.logger.Error(ctx, "error setting cached earliest updated at", logger.Error(err))
	}

	return earliestUpdatedAt, nil
}

func (s *OrchestratorService) generateRequestHash(request *orchestrator_model.GetAggregateRequest) (string, error) {
	hashInput := struct {
		MerchantID                  string   `json:"merchantId"`
		AccountID                   string   `json:"accountId"`
		AccountIDs                  []string `json:"accountIds,omitempty"`
		Statuses                    []string `json:"statuses,omitempty"`
		IncludeFeeIndirectDeduction bool     `json:"includeFeeIndirectDeduction"`
	}{
		MerchantID:                  request.MerchantID.String(),
		AccountID:                   request.AccountID.String(),
		AccountIDs:                  request.AccountIDs,
		Statuses:                    request.Statuses,
		IncludeFeeIndirectDeduction: request.IncludeFeeIndirectDeduction,
	}

	jsonData, err := json.Marshal(hashInput)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonData)
	return base64.URLEncoding.EncodeToString(hash[:])[:16], nil
}
