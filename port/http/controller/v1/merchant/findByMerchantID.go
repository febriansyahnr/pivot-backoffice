package merchant

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// FindByMerchantID		godoc
// @Summary				Find merchant by ID
// @Description			Find merchant by ID
// @ID					merchant-find-by-id
// @Tags				API - Merchant
// @Accept				json
// @Produce				json
// @Param 				merchantID	path		string true "Merchant ID"
// @Success				200  		{object}	response.ApiResponse{data=merchant.MerchantResponse}
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/merchants/:id [get]
// @Security			Bearer
func (c *MerchantController) FindByMerchantID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/FindByMerchantID")
	defer segment.End()

	var (
		err      error
		merchant *merchantModel.Merchant
	)

	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("merchant id is required")))
		return
	}

	if merchant, err = c.merchantSvc.FindMerchantByID(r.Context(), id); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if merchant == nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, err))
		return
	}

	response.SendApiResponseOK(w, merchant.ToResponse())
}
