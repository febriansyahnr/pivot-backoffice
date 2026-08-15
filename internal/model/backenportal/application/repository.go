package application

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	bankAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/bankAccount"
	disbursementRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	merchantRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	snapCoreRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/snapCore"
	statusHistoriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/statusHistories"
)

type AppRepository struct {
	merchantRepository        repository.IMerchantRepository
	disbursementRepository    repository.IDisbursementRepository
	snapCoreRepository        repository.ISnapCoreRepository
	bankAccountRepository     repository.IBankAccountRepository
	statusHistoriesRepository repository.IStatusHistoriesRepository
}

func (a *Application) SetupRepositories() {
	a.repo = AppRepository{}
	a.repo.merchantRepository = merchantRepository.New(a.mySqlDB, a.pdkLog, merchantRepository.WithServiceConfig(a.cfg))
	a.repo.disbursementRepository = disbursementRepository.New(
		a.mySqlDB, a.pdkLog,
		disbursementRepository.WithConfig(&a.cfg.DisbursementConfig),
		disbursementRepository.WithAppConfig(&a.cfg.AppConfig),
	)
	a.repo.snapCoreRepository = snapCoreRepository.New(a.cfg, a.secret, a.pdkLog, a.httpRequestClient)
	a.repo.bankAccountRepository = bankAccountRepository.New(a.mySqlDB, a.pdkLog)
	a.repo.statusHistoriesRepository = statusHistoriesRepository.New(a.mySqlDB)
}

func (a *Application) GetMerchantRepository() repository.IMerchantRepository {
	return a.repo.merchantRepository
}

func (a *Application) GetDisbursementRepository() repository.IDisbursementRepository {
	return a.repo.disbursementRepository
}

func (a *Application) GetSnapCoreRepository() repository.ISnapCoreRepository {
	return a.repo.snapCoreRepository
}

func (a *Application) GetBankAccountRepository() repository.IBankAccountRepository {
	return a.repo.bankAccountRepository
}

func (a *Application) GetStatusHistoriesRepository() repository.IStatusHistoriesRepository {
	return a.repo.statusHistoriesRepository
}
