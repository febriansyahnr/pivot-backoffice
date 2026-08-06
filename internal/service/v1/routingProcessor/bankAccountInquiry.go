package routingprocessorService

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/go/snap"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *routingProcessorService) AccountInquiry(ctx context.Context, request *routingProcessorModel.InquiryAccountRequest) (inquiry *routingProcessorModel.InquiryAccountResponseData, err error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/routingProcessor/AccountInquiry")
	defer span.End()

	for _, routeConfig := range s.GetProcessorList(ctx, request.MerchantID) {
		if !routeConfig.IsActive {
			continue
		}

		if s.routingProcessor == nil {
			s.logger.Error(ctx, "RoutingProcessorService | routingProcessor map is nil")
			continue
		}
		
		processor, exists := s.routingProcessor[routeConfig.ProcessorName]
		if !exists {
			s.logger.Warn(ctx, "RoutingProcessorService | processor not found", 
				logger.String("processor", routeConfig.ProcessorName))
			continue
		}
		
		inquiry, err = processor.BankAccountInquiry(ctx, request)
		if err != nil || (inquiry != nil && !util.IsPatternMatch(constant.SnapCoreResponseCodeSuccessPattern, inquiry.ResponseCode)) {
			if inquiry == nil {
				inquiry = &routingProcessorModel.InquiryAccountResponseData{
					ResponseMessage: err.Error(),
				}
			}

			s.logger.Warn(ctx, "RoutingProcessorService | suspicious response when using inquiry account routing",
				logger.String("processor", routeConfig.ProcessorName),
				logger.Any("Response", inquiry))
		} else {
			break
		}
	}

	if inquiry == nil && err == nil {
		responseCode, responseMessage := snap.GenerateResponseCode(snap.SNAP_INTERNAL_SERVER_ERROR, snap.SNAP_SERVICE_ACCOUNT_INQUIRY_EXTERNAL)
		inquiry = &routingProcessorModel.InquiryAccountResponseData{
			ResponseCode:    responseCode,
			ResponseMessage: responseMessage,
		}

		err = errors.New(responseMessage)
	}

	return inquiry, err
}
