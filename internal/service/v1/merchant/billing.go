package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *MerchantService) GetBillingFees(ctx context.Context, request merchant.BillingFeeRequest) (*merchant.BillingFeeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetBillingFees")
	defer segment.End()

	merchantData, err := s.repo.FindMerchantByID(ctx, request.MerchantId)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if merchantData == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, err)
	}

	billing, err := s.repo.GetBillingFees(ctx, request)
	if err != nil {
		return nil, err
	}

	resp := &merchant.BillingFeeResponse{
		MerchantId:     request.MerchantId,
		MerchantName:   merchantData.Name,
		Total:          0,
		TotalFeeAmount: 0,
		Details:        map[string][]merchant.BillingFeeDetailResponse{},
		SubMerchants:   nil,
	}

	// map of usecase, e.g. payouts and payments to the submerchant id
	subMerchantFees := make(map[string]map[string][]merchant.BillingFeeDetailResponse)
	// map from the submerchant id to the total
	subMerchantTotals := make(map[string]struct {
		Total          int
		TotalFeeAmount float64
	})

	// we need to match between the usecase, e.g. payouts and payments
	// and assign each item to the corresponding merchant id
	for feeUsecase, details := range billing.Details {
		for _, detail := range details {
			// try to match with the main merchant
			if detail.MerchantId == request.MerchantId {
				if resp.Details[feeUsecase] == nil {
					resp.Details[feeUsecase] = []merchant.BillingFeeDetailResponse{}
				}

				resp.Details[feeUsecase] = append(resp.Details[feeUsecase], detail)
				resp.Total += detail.Total
				resp.TotalFeeAmount += detail.TotalFeeAmount
				continue
			}

			// if not, then it is the sub merchant
			if subMerchantFees[detail.MerchantId] == nil {
				subMerchantFees[detail.MerchantId] = make(map[string][]merchant.BillingFeeDetailResponse)
			}

			if subMerchantFees[detail.MerchantId][feeUsecase] == nil {
				subMerchantFees[detail.MerchantId][feeUsecase] = []merchant.BillingFeeDetailResponse{}
			}
			subMerchantFees[detail.MerchantId][feeUsecase] = append(subMerchantFees[detail.MerchantId][feeUsecase], detail)

			totals := subMerchantTotals[detail.MerchantId]
			totals.Total += detail.Total
			totals.TotalFeeAmount += detail.TotalFeeAmount
			subMerchantTotals[detail.MerchantId] = totals
		}
	}

	if len(subMerchantFees) > 0 {
		// fill in the submerchants fields
		resp.SubMerchants = []merchant.SubMerchantBillingResponse{}
		for subMerchantId, details := range subMerchantFees {
			subMerchantData, err := s.repo.FindMerchantByID(ctx, subMerchantId)
			if err != nil || subMerchantData == nil {
				continue
			}

			totals := subMerchantTotals[subMerchantId]
			subMerchantBilling := merchant.SubMerchantBillingResponse{
				SubMerchantId:   subMerchantId,
				SubMerchantName: subMerchantData.Name,
				Total:           totals.Total,
				TotalFeeAmount:  totals.TotalFeeAmount,
				Details:         details,
			}

			resp.SubMerchants = append(resp.SubMerchants, subMerchantBilling)
			resp.Total += totals.Total
			resp.TotalFeeAmount += totals.TotalFeeAmount
		}
	}

	return resp, nil
}

func (s *MerchantService) PayBillingFees(ctx context.Context, request merchant.PayBillingFeeRequest) (*merchant.BillingFeeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/PayBillingFees")
	defer segment.End()

	billingFees, err := s.repo.GetBillingFees(ctx, merchant.BillingFeeRequest{
		MerchantId:              request.MerchantId,
		Status:                  "PENDING",
		BillingDateRangeRequest: request.BillingDateRangeRequest,
	})
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if billingFees.TotalFeeAmount == 0 {
		return billingFees, nil
	}

	if err := s.repo.PayBillingFees(ctx, request); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return billingFees, nil
}
