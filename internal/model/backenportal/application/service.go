package application

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	disbursementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
)

type AppService struct {
	disbursementService service.IDisbursementService

	releaser []func()
}

func (a *Application) SetupServices() {
	disbursementSvc := disbursementService.New(
		a.cfg, a.pdkLog, a.repo.merchantRepository, a.repo.disbursementRepository,
		a.repo.snapCoreRepository, a.repo.bankAccountRepository,
		disbursementService.WithStatusHistoriesRepository(a.repo.statusHistoriesRepository),
	)
	a.service.disbursementService = disbursementSvc
	a.service.releaser = append(a.service.releaser, func() {
		a.service.disbursementService.WPRelease()
	})
}

func (a *Application) ReleaseServices() {
	for _, release := range a.service.releaser {
		release()
	}
}

func (a *Application) GetDisbursementService() service.IDisbursementService {
	return a.service.disbursementService
}
