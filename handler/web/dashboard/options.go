package dashboardhandler

import service "github.com/paper-indonesia/pivot-backoffice/internal/service/backendportal"

type dashboardOption func(*Dashboard)

func WithDisbursementService(disbursementSvc service.IDisbursementService) dashboardOption {
	return func(d *Dashboard) {
		d.disbursementSvc = disbursementSvc
	}
}
