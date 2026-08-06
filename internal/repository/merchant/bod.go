package merchant

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const bodsTableName = "merchant_bods"

func (r *MerchantRepository) ValidateMerchantBODData(ctx context.Context, req *merchant.UpsertBoardOfDirectorReq) (*merchant.BODValidation, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/ValidateMerchantBODData")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, bodsTableName)

	var (
		rawQuery  string
		whereArgs = []string{}
		args      []interface{}
		result    = &merchant.BODValidation{}
	)
	if req.Method == constant.ActionPost {
		result.IsCreate = true
		whereArgs = append(whereArgs, "merchant_id = ? AND position = ? AND NOT deleted")
		args = []interface{}{req.MerchantId, req.Position}
		rawQuery = `SELECT (COUNT(id) = 0) AS valid, '{}' AS identity_file, '' AS hash FROM ` + bodsTableName

		if req.Position == constant.MerchantBODPositionShareholder {
			whereArgs = append(whereArgs, "LOWER(name) = LOWER(?)")
			args = append(args, req.Name)
		} else {
			whereArgs = append(whereArgs, "identity_number = ?")
			args = append(args, req.IdentityNumber)
		}
		rawQuery += ` WHERE ` + strings.Join(whereArgs, " AND ")

	} else {
		args = []interface{}{req.MerchantId, req.Id}
		rawQuery = `SELECT (COUNT(id) = 1) AS valid, IFNULL(identity_file, '{}') AS identity_file, IFNULL(hash, '') AS hash FROM ` + bodsTableName +
			` WHERE merchant_id = ? AND id = ? AND NOT deleted GROUP BY identity_file, hash;`
	}

	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	_ = json.Unmarshal(result.IdentityFile, &result.ObjIdentityFile)
	return result, nil
}

func (r *MerchantRepository) UpsertMerchantBOD(ctx context.Context, action string, data *merchant.BoardOfDirector) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/UpsertMerchantBOD")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, bodsTableName)

	var rawQuery string
	if action == constant.ActionPost {
		rawQuery = `INSERT INTO merchant_bods
			(id, merchant_id, position, name, identity_number, identity_file, position_long, status, shares, created_by, created_at, approved_by, approved_at, updated_at, hash)
		VALUES(:id, :merchant_id, :position, :name, :identity_number, :identity_file, :position_long, :status, :shares, :created_by, :created_at, :approved_by, :approved_at, :updated_at, :hash);`

	} else {
		var setQuery []string
		if data.Position != "" {
			setQuery = append(setQuery, "position = :position")
		}
		if data.Name != "" {
			setQuery = append(setQuery, "name=:name")
		}
		if data.IdentityNumber != "" {
			setQuery = append(setQuery, "identity_number=:identity_number")
		}

		// Use injected marshal func for testing, otherwise use default MarshalJSON
		var b []byte
		if r.jsonMarshalFunc != nil {
			b, _ = r.jsonMarshalFunc(data.IdentityFile)
		} else {
			b, _ = data.IdentityFile.MarshalJSON()
		}
		if len(b) > 0 {
			setQuery = append(setQuery, "identity_file=:identity_file")
		}

		if data.PositionLong != "" {
			setQuery = append(setQuery, "position_long=:position_long")
		}
		if data.Hash != "" {
			setQuery = append(setQuery, "hash=:hash")
		}
		if !data.UpdatedAt.IsZero() {
			setQuery = append(setQuery, "updated_at=:updated_at")
		}
		if data.Shares != nil {
			setQuery = append(setQuery, "shares=:shares")
		}

		if len(setQuery) == 0 {
			r.logger.Warn(ctx, "no update values", logger.Any("data", data))
			return nil
		}
		rawQuery = fmt.Sprintf("UPDATE merchant_bods SET %s WHERE id=:id;", strings.Join(setQuery, ", "))
	}

	_, err := r.db.NamedExecContext(ctx, rawQuery, data)
	return err
}

func (r *MerchantRepository) GetListMerchantBODs(ctx context.Context, merchantId string) (resp []merchant.BoardOfDirectorResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetListMerchantBODs")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, bodsTableName)

	rawQuery := `SELECT id, position, name, identity_number, identity_file, position_long, IFNULL(shares,0) as shares, created_by, created_at, updated_at
		FROM ` + bodsTableName + ` WHERE merchant_id = ? AND NOT deleted ORDER BY position, name`

	if err = r.db.SelectContext(ctx, &resp, rawQuery, merchantId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}
