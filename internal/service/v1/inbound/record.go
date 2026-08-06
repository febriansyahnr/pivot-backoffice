package inboundService

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	inboundPdk "github.com/paper-indonesia/pdk/v2/chiExt/inbound"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func (s *InboundService) Record(ctx context.Context, input inboundPdk.HttpInbound) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/inbound/Record")
	defer span.End()

	if input.Client == nil {
		return nil // Return if it is not inbound client
	}

	id, _ := uuid.NewV7()

	// Insert into db
	request := &inboundModel.InboundRequest{
		ID:                id.String(),
		IP:                input.IP,
		Method:            input.Method,
		URL:               input.URL,
		Headers:           s.inspector.Inspects(input.Headers).(map[string][]string),
		Body:              s.inspector.Inspects(input.Body),
		StatusCode:        input.StatusCode,
		ResponseTimeMs:    input.ResponseTimeMs,
		SnapCompatibility: false,
		Client: &inboundModel.Client{
			Feature:     input.Client.Feature,
			OriginId:    input.Client.OriginId,
			ReferenceId: input.Client.ReferenceId,
			TraceId:     input.TraceId,
		},
	}

	request.ResponseBody, _ = json.Marshal(s.inspector.Inspects(input.ResponseBody))
	request.Client.OriginId = getOriginIDFromResponse(request.ResponseBody) // Force to find originId from response

	if err := s.inboundRepo.Insert(ctx, request); err != nil {
		s.logger.Error(ctx, "failed to insert inbound", pdkLogger.Error(err))
		return err
	}

	return nil
}

func getOriginIDFromResponse(responseBody []byte) (originID string) {
	var tempStruct struct {
		Data struct {
			ID   string `json:"id"`
			UUID string `json:"uuid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(responseBody, &tempStruct); err != nil {
		return
	}

	if tempStruct.Data.ID != "" {
		originID = tempStruct.Data.ID
	} else if tempStruct.Data.UUID != "" {
		originID = tempStruct.Data.UUID
	}

	return
}
