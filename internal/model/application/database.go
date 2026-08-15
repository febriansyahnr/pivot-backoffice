package application

import (
	"fmt"
	"time"

	pdkMySql "github.com/paper-indonesia/pdk/v2/mySqlExt"
	pdkRedis "github.com/paper-indonesia/pdk/v2/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
)

func newDatabase(cfg *config.MySQLConfig, secret *config.MySQLSecret, app *Application) (mySqlExt.IMySqlExt, error) {
	return mySqlExt.New(
		pdkMySql.Config{
			Host:         cfg.Host,
			Port:         cfg.Port,
			Username:     secret.Username,
			Password:     secret.Password,
			DBName:       secret.Database,
			MaxIdleConns: cfg.MaxIdleConns,
			MaxIdleTime:  cfg.MaxOpenConns,
			MaxLifeTime:  cfg.MaxLifeTime,
			MaxOpenConns: cfg.MaxOpenConns,
			SlaveHost:    cfg.SlaveHost,
			SlavePort:    cfg.SlavePort,
		},
		pdkMySql.WithLogger(app.pdkLog),
		pdkMySql.WithTracerProvider(app.otel.TracerProvider()),
		pdkMySql.WithMetricProvider(app.otel.MeterProvider()),
	)
}

func (a *Application) setupDatabase() {
	var err error
	// Init MySql Database
	a.mySqlDB, err = newDatabase(&a.cfg.MySQLConfig, &a.secret.MySQLSecret.Service, a)
	if err != nil {
		fmt.Printf("Unable to init mysql, %v", err)
		panic(err)
	}
	a.AddCloser(func() { a.mySqlDB.Close() })

	// Init Redis
	a.redisClient, err = redisExt.New(
		pdkRedis.Config{
			Addr:             a.cfg.RedisConfig.Host + ":" + a.cfg.RedisConfig.Port,
			Password:         a.secret.RedisSecret.Password,
			DB:               a.cfg.RedisConfig.CacheDB,
			IsRedsyncEnabled: true,
		},
		pdkRedis.WithTracerProvider(a.otel.TracerProvider()),
		pdkRedis.WithMetricProvider(a.otel.MeterProvider()),
		pdkRedis.WithMaxRetries(a.cfg.RedisConfig.MaxRetries),
		pdkRedis.WithMinRetryBackoff(time.Duration(a.cfg.RedisConfig.MinRetryBackoff)*time.Second),
		pdkRedis.WithMaxRetryBackoff(time.Duration(a.cfg.RedisConfig.MaxRetryBackoff)*time.Second),
		pdkRedis.WithDialTimeout(time.Duration(a.cfg.RedisConfig.DialTimeout)*time.Second),
		pdkRedis.WithReadTimeout(time.Duration(a.cfg.RedisConfig.ReadTimeout)*time.Second),
		pdkRedis.WithWriteTimeout(time.Duration(a.cfg.RedisConfig.WriteTimeout)*time.Second),
		pdkRedis.WithPoolSize(a.cfg.RedisConfig.PoolSize),
		pdkRedis.WithPoolTimeout(time.Duration(a.cfg.RedisConfig.PoolTimeout)*time.Second),
	)
	if err != nil {
		fmt.Printf("Unable to init redis cache, %v", err)
		panic(err)
	}
	a.AddCloser(func() { a.redisClient.Close() })
}
