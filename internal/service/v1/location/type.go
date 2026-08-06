package location

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
)

type locationService struct {
	logger logger.ILogger
	repo   repository.IAddrLocationRepository
}

func New(logger logger.ILogger, repo repository.IAddrLocationRepository) service.IAddrLocationService {
	return &locationService{logger, repo}
}
