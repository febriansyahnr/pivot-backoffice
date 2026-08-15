package outbound

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/outbound"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *repository) FindByID(ctx context.Context, id string) (*outbound.Outbound, error) {

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "outbound")

	var result outbound.Outbound
	rawQuery := `SELECT id, client, date, method, url, headers, body, status_code, response_time, response_body, error_message
		FROM outbound
		WHERE id = ?`

	if err := r.db.GetContext(ctx, &result, rawQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &result, nil
}
