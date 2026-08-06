package callbackController

import (
	"net/http"

	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CallbackController) GetCallbackEvents(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/callback/GetCallbackEvents")
	defer segment.End()

	events, err := c.callbackSvc.GetCallbackEvents(ctx)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponseOK(w, events)
}
