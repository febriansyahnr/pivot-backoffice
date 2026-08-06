package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
)

func (s *MerchantService) GetNotificationConfig(ctx context.Context, merchantID string) (*merchant.MerchantNotificationConfig, error) {
	m, err := s.repo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, constant.ErrMerchantNotFound
	}

	metadata, err := m.GetMetadata()
	if err != nil {
		return nil, err
	}

	if metadata == nil || metadata.NotificationConfig == nil {
		return &merchant.MerchantNotificationConfig{}, nil
	}

	return metadata.NotificationConfig, nil
}

func (s *MerchantService) UpdateNotificationConfig(ctx context.Context, merchantID string, req *merchant.MerchantNotificationConfig) (*merchant.MerchantNotificationConfig, error) {
	merchant, err := s.repo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if merchant == nil {
		return nil, constant.ErrMerchantNotFound
	}

	if err := merchant.UpdateNotificationConfig(req); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, merchant); err != nil {
		return nil, err
	}

	return s.GetNotificationConfig(ctx, merchantID)
}
