package application

import (
	beRepo "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	bankAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/bankAccount"
	disbursementRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/disbursement"
	merchantRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/merchant"
	snapCoreRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/snapCore"
	statusHistoriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/statusHistories"
)

type AppRepository struct {
	merchantRepository        beRepo.IMerchantRepository
	disbursementRepository    beRepo.IDisbursementRepository
	snapCoreRepository        beRepo.ISnapCoreRepository
	bankAccountRepository     beRepo.IBankAccountRepository
	statusHistoriesRepository beRepo.IStatusHistoriesRepository
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
