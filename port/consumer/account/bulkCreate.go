package account

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/paper-indonesia/pdk/v2/logger"

	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *AccountConsumer) BulkCreateAccount(ctx context.Context, body []byte, channel string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/account/BulkCreateAccount")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from BulkCreateAccount", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var req account_model.BulkCreateAccountRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		return pkgError.New(response.HttpErrRequest, err)
	}

	err = c.AccountSvc.BulkCreateAccount(ctx, &req)
	if err != nil {
		return err
	}
	return nil
}
