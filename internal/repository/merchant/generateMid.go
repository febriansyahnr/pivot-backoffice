package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MerchantRepository) GenerateNewMID(ctx context.Context) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GenerateNewMID")
	defer segment.End()

	var mid string

	query := `
		SELECT m.mid FROM merchants as m ORDER by m.mid DESC LIMIT 1`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantsTable)

	if err := r.db.GetContext(ctx, &mid, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "0000", nil
		}

		r.logger.Error(ctx, "error get latest mid", logger.Error(err))
		return "", err
	}

	// Generate MID by latest mid + 1
	midInInteger, err := strconv.Atoi(mid)
	if err != nil {
		r.logger.Error(ctx, "error when generate new MID", logger.Error(err))
		return "", nil
	}

	return fmt.Sprintf("%04s", strconv.Itoa(midInInteger+1)), nil
}
