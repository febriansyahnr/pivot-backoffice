package beneficiaryAccountController

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetList godoc
// @Summary			Get beneficiary accounts list
// @Description		Get beneficiary accounts list
// @ID				api-beneficiary-account-get-list
// @Tags			API - Beneficiary Account
// @Accept			json
// @Produce			json
// @Param        	page    				query     	string  false  "pagination current page"
// @Param        	beneficiaryAccountNo    query     	string  false  "filter beneficiaryAccountNo"
// @Param        	beneficiaryAccountName  query     	string  false  "filter beneficiaryAccountName"
// @Param       	startCreatedAt			query     	string  false  "filter startCreatedAt"
// @Param       	endCreatedAt			query     	string  false  "filter endCreatedAt"
// @Success			200  					{object}	response.ApiResponse{data=[]beneficiaryAccountModel.BeneficiaryAccount,meta=commonModel.Meta}
// @Failure			500  					{object}	response.ApiResponse
// @Router			/api/v1/beneficiary-accounts [get]
// @Security		Bearer
func (c *Controller) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/beneficiaryAccount/GetList")
	defer segment.End()

	var (
		startCreatedAt         *time.Time // default nil
		endCreatedAt           *time.Time // default nil
		beneficiaryAccountNo   string     // default nil
		beneficiaryAccountName string     // default nil
		isXb                              = false
		page                   int64      = 1
		err                    error
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Get query params
	pageStr := r.URL.Query().Get("page")
	startCreatedAtStr := r.URL.Query().Get("startCreatedAt")
	endCreatedAtStr := r.URL.Query().Get("endCreatedAt")
	beneficiaryAccountNo = r.URL.Query().Get("beneficiaryAccountNo")
	beneficiaryAccountName = r.URL.Query().Get("beneficiaryAccountName")
	isXbStr := r.URL.Query().Get("isXb")

	// Validation and parsing
	if startCreatedAtStr != "" {
		parsedStartCreatedAt, err := time.Parse(util.UTCLayout, startCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid startCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		startCreatedAt = &parsedStartCreatedAt
	}
	if endCreatedAtStr != "" {
		parsedEndCreatedAt, err := time.Parse(util.UTCLayout, endCreatedAtStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest,
				fmt.Errorf("invalid endCreatedAt format. Use 'YYYY-MM-DDTHH:mm:ssZ' format")))
			return
		}

		endCreatedAt = &parsedEndCreatedAt
	}
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}
	if isXbStr != "" {
		isXb, err = strconv.ParseBool(isXbStr)
		if err != nil {
			response.SendApiResponseError(ctx, w, errors.New(
				response.HttpErrRequest, fmt.Errorf("invalid isXb format. Use bool format instead")))
			return
		}
	}

	filter := &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{
		MerchantID:             user.MerchantId,
		StartCreatedAt:         startCreatedAt,
		EndCreatedAt:           endCreatedAt,
		BeneficiaryAccountNo:   beneficiaryAccountNo,
		BeneficiaryAccountName: beneficiaryAccountName,
		IsXb:                   isXb,
	}
	var perPage int64 = constant.DefaultPaginationPageSize
	if c.config != nil {
		perPage = c.config.AppConfig.PaginationPerPage
	}

	list, err := c.beneficiaryAccountSvc.GetList(r.Context(), filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, err))
		return
	}

	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}
