package reconciliation

import (
	"encoding/json"
	"net/http"

	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *ReconciliationController) DownloadResult(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/reconciliation/DownloadResult")
	defer segment.End()

	type Request struct {
		UUID string `json:"uuid"`
	}
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	fileUrl, err := c.reconSvc.DownloadResult(ctx, request.UUID)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponseOK(w, map[string]string{"result_url": fileUrl})
}
