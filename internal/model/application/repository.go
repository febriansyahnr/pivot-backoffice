package application

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	bankAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/bankAccount"
	disbursementRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	merchantRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	snapCoreRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/snapCore"
	statusHistoriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/statusHistories"
)

func (a *Application) GetMerchantRepository() repository.IMerchantRepository {
	return merchantRepository.New(a.mySqlDB, a.pdkLog, merchantRepository.WithServiceConfig(a.cfg))
}

func (a *Application) GetDisbursementRepository() repository.IDisbursementRepository {
	return disbursementRepository.New(a.mySqlDB, a.pdkLog,
		disbursementRepository.WithConfig(&a.cfg.DisbursementConfig),
		disbursementRepository.WithAppConfig(&a.cfg.AppConfig),
	)
}

func (a *Application) GetSnapCoreRepository() repository.ISnapCoreRepository {
	return snapCoreRepository.New(a.cfg, a.secret, a.pdkLog, a.httpRequestClient)
}

func (a *Application) GetBankAccountRepository() repository.IBankAccountRepository {
	return bankAccountRepository.New(a.mySqlDB, a.pdkLog)
}

func (a *Application) GetStatusHistoriesRepository() repository.IStatusHistoriesRepository {
	return statusHistoriesRepository.New(a.mySqlDB)
}
