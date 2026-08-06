package withdrawalController

import (
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

type handler struct {
	validator *validatorExt.Validate
	logger    logger.ILogger
	service   service.IWithdrawalService
	userSvc   service.IUserService
}

var (
	loc, _     = time.LoadLocation(constant.TimeLoc)
	otelTracer = otel.Tracer("WithdrawalController")

	pathAccountNames = []string{"payments", "payouts"}
)

func New(vld *validatorExt.Validate, log logger.ILogger, svc service.IWithdrawalService, userSvc service.IUserService) controller.V1WithdrawalController {
	return &handler{
		validator: vld, logger: log, service: svc, userSvc: userSvc,
	}
}

func (h *handler) preparationGetList(r *http.Request, account string, request *withdrawal.WithdrawalListRequest) (err error) {
	if account == "payments" {
		request.AccountName = constant.TypePayment
	}

	if request.StartDate, err = time.ParseInLocation(time.DateOnly, request.StrStartDate, loc); err != nil {
		return pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidStartDateFmt)
	}
	request.StartDate = request.StartDate.UTC()

	if request.EndDate, err = time.ParseInLocation(time.DateOnly, request.StrEndDate, loc); err != nil {
		return pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidEndDateFmt)
	}
	request.EndDate = request.EndDate.UTC().Add((24 * time.Hour) - time.Millisecond)

	params := r.URL.Query() // For update query params
	params.Set("startDate", request.StartDate.Format(time.RFC3339))
	params.Set("endDate", request.EndDate.Format(time.RFC3339))

	r.URL.RawQuery = params.Encode()

	if err = httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate"); err != nil {
		return pkgErrs.New(response.HttpErrRequest, err)
	}
	return nil
}
