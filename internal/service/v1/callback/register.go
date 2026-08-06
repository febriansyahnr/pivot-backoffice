package callbackService

import (
	"context"

	"github.com/google/uuid"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *CallbackService) RegisterCallback(
	ctx context.Context,
	request *callbackModel.RegisterCallbackRequest,
) (*callbackModel.RegisterCallbackResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/RegisterCallback")
	defer segment.End()

	// Find Callback Master First
	callbackMaster, err := s.callbackRepo.FindCallbackMasterByName(
		ctx,
		util.ToTitle(request.Name))
	if err != nil {
		return nil, errors.New(httpResponse.HttpErrDatabase, err)
	}

	var callbackMasterUUID uuid.UUID
	callbackMasterForResponse := request.ToCallbackMaster()
	// Checking Callback Master
	if callbackMaster != nil {
		callbackMasterUUID = callbackMaster.UUID
		callbackMasterForResponse = callbackMaster
	} else {
		// Create Callback Master First
		callbackMasterUUID = callbackMasterForResponse.UUID
		if err = s.callbackRepo.CreateCallbackMaster(ctx, callbackMasterForResponse); err != nil {
			return nil, errors.New(httpResponse.HttpErrDatabase, err)
		}
	}

	// Create Callback based on Callback Master
	callback := request.ToCallback(callbackMasterUUID)
	if err = s.callbackRepo.CreateCallback(ctx, callback); err != nil {
		return nil, errors.New(httpResponse.HttpErrDatabase, err)
	}

	return request.ToResponse(callbackMasterForResponse, callback.UUID), nil
}
