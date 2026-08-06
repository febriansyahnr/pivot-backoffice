package accountController

import (
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (b *account) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/account/GetBalance")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	query := r.URL.Query()

	currency := query.Get("currency")
	if currency != "" && currency != constant.CurrencyIDR {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrAccountNotFound))
		return
	}

	usecaseType := query.Get("usecase")
	if usecaseType == "" {
		// default usecase is disbursement
		usecaseType = constant.TypeDisbursement
	}

	request := orchestratorModel.GetMerchantBalanceRequest{
		Date:        time.Now().UTC(),
		MerchantID:  user.MerchantId,
		BalanceName: accountModel.GetAccountNameByUsecase(usecaseType),
	}

	result, err := b.orchestratorSvc.GetMerchantBalance(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, result)
}
