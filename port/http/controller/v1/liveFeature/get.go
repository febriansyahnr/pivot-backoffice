package liveFeature

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *LiveFeatureController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/liveFeature/GetList")
	defer segment.End()

	var (
		err error
	)

	res, err := c.featureSvc.GetList(ctx)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, res)
}

func (c *LiveFeatureController) GetAppVersion(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/liveFeature/GetAppVersion")
	defer segment.End()

	images, err := c.featureSvc.GetAppVersion(ctx)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, images)
}
