package orchestratorController

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"

	"github.com/go-playground/validator/v10"
)

const (
	maxDateRangeInDays = 31
	maxBackdateInDays  = 180
	maxPaginationPage  = 200
	minPaginationValue = 1
)

var otelTracer = otel.Tracer("OrchestratorController")

type OrchestratorController struct {
	config          *config.Config
	orchestratorSvc service.IOrchestratorService
	merchantSvc     service.IMerchantService
	reportingSvc    service.IReportingService
	validate        *validator.Validate
}

func New(config *config.Config, orchestratorSvc service.IOrchestratorService, merchantSvc service.IMerchantService, validate *validator.Validate, reportingSvc service.IReportingService) controller.V1OrchestratorController {
	return &OrchestratorController{
		config:          config,
		orchestratorSvc: orchestratorSvc,
		merchantSvc:     merchantSvc,
		validate:        validate,
		reportingSvc:    reportingSvc,
	}
}

func (o *OrchestratorController) getQueryForTransactionHistory(r *http.Request) (filter *orchestratorModel.TransactionHistoryFilterRequest, err error) {

	user, ok := r.Context().Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		return nil, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
	}

	filter = &orchestratorModel.TransactionHistoryFilterRequest{
		MerchantID:          user.MerchantId,
		TrxTypes:            r.URL.Query()["trxTypes"],
		Status:              r.URL.Query().Get("status"),
		TrxID:               r.URL.Query().Get("trxId"),
		TransactionId:       r.URL.Query().Get("id"),
		BalanceTypes:        r.URL.Query()["balanceTypes"],
		MerchantReferenceID: r.URL.Query().Get("merchantReferenceId"),
		SettlementModel:     constant.PaymentMethodChannelTypeAggregator,
	}

	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sort == "" {
		sort = "-date"
	}
	if filter.FilteredSortQuery, ok = sortFieldForGetList(sort); !ok {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New("invalid data sorting column"))
	}
	startSettlementDate := r.URL.Query().Get("startSettlementDate")
	endSettlementDate := r.URL.Query().Get("endSettlementDate")

	if startSettlementDate == "" || endSettlementDate == "" {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New("start settlement date and end settlement date must be filled"))
	}

	if filter.StartSettlementDate, err = time.ParseInLocation(constant.DateFormat, startSettlementDate, loc); err != nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidStartDateFmt)
	}
	filter.StartSettlementDate = filter.StartSettlementDate.UTC()

	if filter.EndSettlementDate, err = time.ParseInLocation(constant.DateFormat, endSettlementDate, loc); err != nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidEndDateFmt)
	}
	filter.EndSettlementDate = filter.EndSettlementDate.UTC().Add((24 * time.Hour) - time.Millisecond)

	if filter.StartSettlementDate.After(filter.EndSettlementDate) {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrFilterDateInput)

	} else if (filter.EndSettlementDate.Sub(filter.StartSettlementDate).Hours() / 24) > 31 {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New("maximum date range 31 days"))

	} else if (time.Now().UTC().Sub(filter.StartSettlementDate).Hours() / 24) > 180 {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New("maximum backdate 180 days"))
	}

	// Set created_at filter (settlement date - 7 days) for query optimization
	filter.CreatedAtStartDate = filter.StartSettlementDate.AddDate(0, 0, -7)
	filter.CreatedAtEndDate = filter.EndSettlementDate

	// By default, to use AGGREGATOR
	if r.URL.Query().Get("settlementModel") != "" {
		filter.SettlementModel = r.URL.Query().Get("settlementModel")
	}

	return
}

var loc, _ = time.LoadLocation(constant.TimeLoc)

func sortFieldForGetList(key string) (val string, ok bool) {
	val, ok = map[string]string{
		"-date":                   "t.updated_at DESC",
		"date":                    "t.updated_at",
		"-beneficiaryAccountName": "d.beneficiary_account_name DESC",
		"beneficiaryAccountName":  "d.beneficiary_account_name",
		"-amount":                 "((-1*t.debit)+t.credit) DESC",
		"amount":                  "((-1*t.debit)+t.credit)",
	}[key]
	return
}

func sortFieldForGetListViaDataReporting(key string) (val string, ok bool) {
	val, ok = map[string]string{
		"-date":                   "updated_at DESC",
		"date":                    "updated_at",
		"-beneficiaryAccountName": "beneficiary_account_name DESC",
		"beneficiaryAccountName":  "beneficiary_account_name",
		"-amount":                 "amount DESC",
		"amount":                  "amount",
	}[key]
	return
}
