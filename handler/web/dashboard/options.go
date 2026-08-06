package dashboardhandler

import "github.com/paper-indonesia/pivot-backoffice/internal/service"

type dashboardOption func(*Dashboard)

func WithDisbursementService(disbursementSvc service.IDisbursementService) dashboardOption {
	return func(d *Dashboard) {
		d.disbursementSvc = disbursementSvc
	}
}
