package routingprocessorService

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	processorPriorityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/processorPriority"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/go/snap"
	"github.com/paper-indonesia/pdk/v2/logger"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// TriggerTransfer is a function to process bank transfer request. This function is selecting the available routing processor
// based on the flag "backend-portal-processor-routing". The function will iterate all the available routing processor
// and trigger the bank transfer process. If the bank transfer process is failed, the function will log the error
// and return the last error.
func (s *routingProcessorService) BankTransfer(
	ctx context.Context,
	request *routingProcessorModel.BankTransferRequest,
	routeConfigs ...processorPriorityModel.ProcessorPriority,
) (*routingProcessorModel.BankTransferResponseData, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/routingProcessor/TriggerTransfer")
	defer span.End()

	var (
		transfer *routingProcessorModel.BankTransferResponseData
		err      error
	)
	// select priority routing processor fron bank transfer
	processorSelected, _ := ctx.Value(constant.CtxProcessorName).(string)

	if len(routeConfigs) == 0 {
		routeConfigs = s.GetProcessorList(ctx, request.HeaderRequest.MerchantId)
	}

	for _, routeConfig := range routeConfigs {
		if !routeConfig.IsActive {
			continue
		}

		if processorSelected != "" {
			if routeConfig.ProcessorName != processorSelected {
				continue
			}
		}

		// check allowed destination
		if !s.isChannelCodeAllowed(routeConfig.AllowedDestinations, request.Beneficiary.BankCode) {
			s.logger.Warn(ctx, "RoutingProcessorService | processor not eligible for destination",
				logger.String("processor", routeConfig.ProcessorName),
				logger.String("channelMethod", request.Beneficiary.BankCode))

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
		
		transfer, err = processor.TriggerTransfer(ctx, request)
		if err != nil || (transfer != nil &&
			!util.IsPatternMatch(constant.SnapCoreResponseCodeSuccessPattern, transfer.ResponseCode)) {

			if transfer == nil {
				transfer = &routingProcessorModel.BankTransferResponseData{
					ResponseMessage: err.Error(),
				}
			}

			// action need?
			s.logger.Warn(ctx, "RoutingProcessorService | suspicious response when using transfer routing",
				logger.String("processor", routeConfig.ProcessorName),
				logger.Any("Response", transfer))
		}

		if s.isBreakingProcess(transfer, err) {
			break
		}
	}

	if transfer == nil && err == nil {
		responseCode, responseMessage := snap.GenerateResponseCode(snap.SNAP_INTERNAL_SERVER_ERROR, snap.SNAP_SERVICE_INTERBANK_TRANSFER)
		transfer = &routingProcessorModel.BankTransferResponseData{
			ResponseCode:    responseCode,
			ResponseMessage: responseMessage,
		}

		err = errors.New(responseMessage)
	}

	return transfer, err
}

// getProcessorList is a function to get the list of routing processor based on the flag "backend-portal-processor-routing".
// If the flag is not set or error, it will return the default processor which is SNAP_CORE_PROCESSOR.
// The list of processor will be sorted by priority.
func (s *routingProcessorService) GetProcessorList(ctx context.Context, merchantID string) []processorPriorityModel.ProcessorPriority {
	defaultProcessor := []processorPriorityModel.ProcessorPriority{
		{
			ProcessorName: constant.SnapCoreProcessor,
			Priority:      1,
			IsActive:      true,
		},
	}

	parameter := s.cfg.Environment
	if merchantID != "" {
		parameter = merchantID
	}

	processorFlag := ffcontext.NewEvaluationContext(parameter)
	getFF, err := ffclient.JSONArrayVariation("backend-portal-processor-routing", processorFlag, nil)
	if err != nil || getFF == nil {
		s.logger.Error(ctx, "RoutingProcessorService | error when get processor list", logger.Error(err))
		return defaultProcessor
	}

	processors := make([]processorPriorityModel.ProcessorPriority, len(getFF))
	for i, ff := range getFF {
		processor := ff.(map[string]interface{})

		allowedDestinations := make([]string, 0)
		if processor["allowedDestinations"] != nil {
			b, _ := json.Marshal(processor["allowedDestinations"])
			err := json.Unmarshal(b, &allowedDestinations)
			if err != nil {
				s.logger.Error(ctx, "RoutingProcessorService | error when get processor list", logger.Error(err))
				return defaultProcessor
			}
		}

		processors[i] = processorPriorityModel.ProcessorPriority{
			ProcessorName:       processor["processorName"].(string),
			Priority:            processor["priority"].(int),
			IsActive:            processor["isActive"].(bool),
			AllowedDestinations: allowedDestinations,
		}
	}

	sort.Slice(processors, func(i, j int) bool {
		return processors[i].Priority < processors[j].Priority
	})

	return processors
}

func (s *routingProcessorService) isBreakingProcess(responseData *routingProcessorModel.BankTransferResponseData, err error) bool {
	if err != nil && errors.Is(err, constant.ErrDoubleDisbursementIndication) {
		return true

	} else if responseData == nil {
		return false
	}

	return responseData.Status == "SUCCESS" ||
		responseData.Status == "PENDING" ||
		util.IsPatternMatch(constant.SnapCoreResponseCodeRequestInProgress, responseData.ResponseCode)
}

func (s *routingProcessorService) isChannelCodeAllowed(allowedDestinations []string, channelCode string) bool {
	if len(allowedDestinations) == 0 {
		return true
	}

	hasNegative := strings.Contains(allowedDestinations[0], "!")
	for _, destination := range allowedDestinations {
		if strings.Contains(destination, channelCode) && !hasNegative {
			return true
		} else if strings.Contains(destination, channelCode) && hasNegative {
			return false
		}
	}

	return hasNegative
}
