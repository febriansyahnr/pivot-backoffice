package paperCommunication

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paperCommModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paperCommunication"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *paperCommunicationRepository) SendEmailV1(ctx context.Context, from string, data *paperCommModel.Email) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/paperCommunication/SendEmailV1")
	defer segment.End()

	url := r.config.BaseURL + "/v1/email/send"

	r.log.Info(ctx, "Request send email", logger.String("url", url), logger.String("from", from), logger.Any("body", data))

	ctxWT, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	response, statusCode, err := r.httpReq.POST(
		ctxWT, url, data, map[string]string{
			constant.HeaderXCustomFrom:   from,
			constant.HeaderXSenderOrigin: r.config.SenderOrigin,
		},
	)
	if err != nil {
		r.log.Error(ctx, "error when do request send email", logger.Error(err))
		return err

	} else if statusCode != http.StatusOK {
		r.log.Error(ctx, fmt.Sprintf("got error %d when do request send email", statusCode), logger.ByteString("response", response))
		return errors.New("failed to send email, see log for more details")
	}
	return nil
}
