package disbursementConsumerController

import (
	"context"
	"encoding/base64"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/disbursement"

	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func (c *DisbursementConsumer) BatchProcessDisbursement(ctx context.Context, body []byte, _ string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/disbursement/BatchProcessDisbursement")
	defer segment.End()

	payload := &pb.BatchProcessDisbursementRequest{}
	if err = proto.Unmarshal(body, payload); err != nil {
		return constant.ErrUnmarshalProto
	}
	c.logger.Info(
		ctx, "Incoming message on BatchProcessDisbursement",
		logger.String("base64Payload", base64.StdEncoding.EncodeToString(body)), logger.Any("payload", payload),
	)
	request := &disbursementModel.BatchProcessDisbursementRequest{
		BulkID:          payload.BulkId,
		DisbursementIDs: payload.DisbursementIds,
	}

	// Call batch process disbursement
	if err = c.disbursementSvc.BatchProcessDisbursement(ctx, request); err != nil {
		return err
	}
	return nil
}
