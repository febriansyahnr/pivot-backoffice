package unifiedPaymentService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) GetSessionDetail(ctx context.Context, request *unifiedPaymentModel.GetUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetSessionDetail")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentSessionID)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	} else if payment.MerchantID != request.MerchantID {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)
	}

	charge, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	inquiryRequest := unifiedPaymentModel.PerformInquiryRequest{}
	if charge != nil {
		inquiryRequest.LedgerID = charge.UUID.String()
	}
	inquiryResult, err := s.performPaymentInquiry(ctx, payment, inquiryRequest)
	if err != nil {
		s.logger.Warn(ctx, "Payment inquiry failed, returning stored status",
			logger.String("paymentUUID", payment.UUID),
			logger.Error(err),
		)
	}

	if inquiryResult != nil && inquiryResult.UpdatedStatus {
		chargeID := ""
		if charge != nil {
			chargeID = charge.UUID.String()
		}

		if err := s.processNotificationFromInquiry(ctx, payment, inquiryResult, chargeID); err == nil {
			payment, _ = s.paymentRepo.GetPaymentById(ctx, request.PaymentSessionID)
			charge, _ = s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
		}
	}

	if inquiryResult != nil && charge != nil && charge.ProcessorTransactionId == "" {
		updateReq := orchestratorModel.UpdatePaymentTransactionRequest{
			LedgerId:               charge.UUID.String(),
			ProcessorTransactionId: inquiryResult.ProcessorTransactionID,
		}
		if err := s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(ctx, updateReq, orchestratorModel.MetadataPayment[any]{}); err != nil {
			s.logger.Warn(ctx, "failed to update processor_transaction_id from inquiry",
				logger.String("chargeUUID", charge.UUID.String()),
				logger.String("processorTransactionID", inquiryResult.ProcessorTransactionID),
				logger.Error(err),
			)
		} else {
			charge.ProcessorTransactionId = inquiryResult.ProcessorTransactionID
		}
	}

	unifiedPaymentResp := payment.ToUnifiedPaymentAndChargeResponse(charge)

	if inquiryResult != nil && inquiryResult.LastInquiryAt != nil {
		unifiedPaymentResp.LastInquiryAt = inquiryResult.LastInquiryAt
	}

	if payment.CustomerID != "" {
		customer, err := s.customerRepo.GetCustomerById(ctx, payment.CustomerID, payment.MerchantID)
		if err != nil {
			return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
		}

		if customer != nil {
			customer.SetUnifiedPaymentCustomerInfo(unifiedPaymentResp)
		}
	}

	if unifiedPaymentResp.AutoSplitPayment != nil {
		detail, err := s.internalUnifiedPaymentSvc.GetAutoSplitPaymentDetail(ctx, &paymentModel.GetAutoSplitPaymentSummaryRequest{
			ReferenceID:     payment.UUID,
			MerchantID:      payment.MerchantID,
			MaxDateCreation: maxPaymentCreatedDays,
		})
		if err != nil {
			return nil, err
		}

		if detail != nil {
			unifiedPaymentResp.AutoSplitDetails = detail.ToAutoSplitDetail()
		}
	}

	return unifiedPaymentResp, nil
}

func (s *UnifiedPaymentService) GetSessionList(ctx context.Context, request *paymentModel.GetListFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetSessionList")
	defer segment.End()

	// Get payment session by id
	list, err := s.paymentRepo.GetList(ctx, request)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	if listData, ok := list.Data.([]*paymentModel.Payment); ok {
		parsedResponse := make([]*unifiedPaymentModel.UnifiedPaymentSessionResponse, 0, len(listData))
		for _, item := range listData {
			unifiedPaymentResponse := item.ToUnifiedPaymentResponse()
			if unifiedPaymentResponse.IsStaticPayment() {
				charges, err := s.GetChargeList(ctx, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID:       request.MerchantID,
					PaymentSessionID: item.UUID,
					Page:             0,
					PerPage:          1000, // TODO: Handle limit static charges
				})
				if err != nil {
					return nil, pkgErr.New(response.HttpErrDatabase, err)
				}

				if listChargeData, ok := charges.Data.([]*unifiedPaymentModel.ChargeResponse); ok {
					unifiedPaymentResponse.ChargeDetails = append(unifiedPaymentResponse.ChargeDetails, listChargeData...)
				}

				parsedResponse = append(parsedResponse, unifiedPaymentResponse)
			} else {
				// TODO: Next improvement is not using N+1 query
				charge, err := s.accountTransactionRepo.FindByReference(ctx, item.UUID, constant.TypePayment)
				if err != nil {
					return nil, pkgErr.New(response.HttpErrDatabase, err)
				}

				unifiedPaymentResp := item.ToUnifiedPaymentAndChargeResponse(charge)
				if item.CustomerID != "" {
					customer, err := s.customerRepo.GetCustomerById(ctx, item.CustomerID, item.MerchantID)
					if err != nil {
						return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
					}

					if customer != nil {
						customer.SetUnifiedPaymentCustomerInfo(unifiedPaymentResp)
					}
				}

				parsedResponse = append(parsedResponse, unifiedPaymentResp)
			}
		}
		list.Data = parsedResponse
	}

	return list, nil
}

func (s *UnifiedPaymentService) GetCardBinDetail(ctx context.Context, request unifiedPaymentModel.GetBinDetailRequest) (*unifiedPaymentModel.GetBinDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetCardBinDetail")
	defer segment.End()

	paymentMethod, err := s.paymentMethodRepo.GetActivePaymentMethodByRequest(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: request.MerchantId,
		Category:   paymentConst.PAYMENT_METHOD_CATEGORY_PAYMENT,
		Type:       paymentConst.PAYMENT_METHOD_CREDIT_CARD,
	})
	if err != nil {
		s.logger.Error(ctx, "Failed while get active payment method card", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if paymentMethod == nil {
		return nil, constant.NewErrStringRequest(response.HttpErrForbidden, constant.ErrCodeForbiddenAccess, "Merchant not authorized for BIN lookup")
	}

	binDetail, err := s.creditCardProcessorRepo.GetBinDetailByBinNumber(ctx, request.MerchantId, request.BinNumber)
	if err != nil {
		return nil, err

	} else if binDetail == nil {
		return nil, constant.NewErrStringRequest(response.HttpErrNotFound, constant.ErrCodeDataNotFound, "BIN not found")
	}

	return &unifiedPaymentModel.GetBinDetailResponse{
		BIN:       binDetail.BinNumber,
		CardType:  binDetail.CardType,
		Principal: binDetail.CardBrand,
		CardLevel: binDetail.CardLevel,
		Issuer:    binDetail.IssuerName,
		Country:   binDetail.IssuerCountry,
		Currency:  binDetail.Currency,
	}, nil
}

// Helper methods for status history
type statusInfo string

const (
	statusInfoLabel          statusInfo = "label"
	statusInfoDescription               = "description"
	statusInfoRecommendation            = "recommendation"
)

func (*UnifiedPaymentService) getPaymentStatusHistoryLabelsAndDescriptions(status string, label statusInfo) string {
	if statusInfo, exists := constant.PaymentStatusHistoryLabelsAndDescriptions[status]; exists {
		if label, ok := statusInfo[string(label)]; ok {
			return label
		}
	}
	return status
}
