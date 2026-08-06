package activityRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mongoDbExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

type RepositoryType int

var otelTracer = otel.Tracer("ActivityRepository")

const (
	MySQLType RepositoryType = iota
	MongoDBType
)

type ActivityRepository struct {
	Mongo  mongoDbExt.IMongoDbExt
	Mysql  mySqlExt.IMySqlExt
	Logger logger.ILogger
}

func (f *ActivityRepository) CreateRepository(repositoryType RepositoryType) repository.IActivityRepository {
	switch repositoryType {
	case MongoDBType:
		return &MongoDBRepository{mongo: f.Mongo}
	default:
		return &MySqlDBRepository{mysql: f.Mysql, logger: f.Logger}
	}
}
