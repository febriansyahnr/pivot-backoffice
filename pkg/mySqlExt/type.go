package mySqlExt

import (
	"context"
	"database/sql"

	pdkMySql "github.com/paper-indonesia/pdk/v2/mySqlExt"
)

type IMySqlExt interface {
	pdkMySql.IMySqlExt

	ExecContextReturnLastId(
		ctx context.Context,
		query string,
		args ...interface{},
	) (*sql.Result, error)

	// Helper
	Rebind(query string) string
	In(rawQuery string, args ...interface{}) (string, []interface{}, error)
}

type IMySqlRows interface {
	Next() bool
	Close() error
	Scan(dest ...any) error
}

type mySqlExt struct {
	pdkMySql.IMySqlExt
}

func New(config pdkMySql.Config, opts ...pdkMySql.OptionFunc) (IMySqlExt, error) {
	pdkIMySql, err := pdkMySql.New(config, opts...)
	if err != nil {
		return nil, err
	}

	return &mySqlExt{pdkIMySql}, nil
}
