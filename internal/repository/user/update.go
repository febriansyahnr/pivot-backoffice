package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *UserRepository) Update(ctx context.Context, user *userModel.User) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
			UPDATE
				users
			SET name = ?, email = ?, status = ?, password = ?, blocked_at = ?, refresh_token = ?, merchant_id = ?,
			    is_change_password = ?, pin_hash = ?, last_login_at = ?, deactivate_at = ?, updated_at = ?
			WHERE
			    uuid = ?`
	_, err := r.db.ExecContext(
		ctx, query,
		user.Name,
		user.Email,
		user.Status,
		user.Password,
		user.Blocked,
		user.RefreshToken,
		user.MerchantId,
		user.IsChangePassword,
		user.PinHash,
		user.LastLoginAt,
		user.DeactivatedAt,
		user.UpdatedAt,
		user.UUID,
	)
	if err != nil {
		r.logger.Error(ctx, "error when updating user", logger.Error(err))
		return err
	}

	return nil
}

func (r *UserRepository) UpdateRefreshToken(ctx context.Context, id, token string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/UpdateRefreshToken")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `UPDATE users SET refresh_token = ? where uuid = ?`
	_, err := r.db.ExecContext(ctx, query, token, id)
	if err != nil {
		r.logger.Error(ctx, "error when updating user refresh token", logger.Error(err))
		return err
	}

	return nil
}

func (r *UserRepository) ChangePassword(ctx context.Context, id string, password string) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/ChangePassword")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `UPDATE users set password = ? where uuid = ?`
	affected, err := r.db.ExecContext(ctx, query, password, id)
	if err != nil {
		r.logger.Error(ctx, "error when updating user password", logger.Error(err))
		return false, err
	}

	return affected, nil

}

func (r *UserRepository) BlockUser(ctx context.Context, id string, blocked sql.NullTime) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/user/BlockUser")
	defer segment.End()

	_, err = r.db.ExecContext(
		context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName), `UPDATE users SET blocked_at = ?, updated_at = NOW() WHERE uuid = ?`,
		blocked, id,
	)
	return
}

func (r *UserRepository) UpdateUserTOTPData(ctx context.Context, request *userModel.UpdateUserTOTPDataRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/user/UpdateUserTOTPData")
	defer segment.End()

	if request == nil {
		return errors.New("request parameters can't be nil")
	}

	request.UpdatedAt = time.Now().UTC()

	fields := []string{
		"updated_at = :updated_at",
	}
	if request.WrappedSecret != "" {
		fields = append(fields, "totp_wrapped_secret = :totp_wrapped_secret")
	}
	if request.EncryptVersion > 0 {
		fields = append(fields, "totp_encrypt_version = :totp_encrypt_version")
	}
	if request.Status != "" {
		fields = append(fields, "totp_status = :totp_status")
	}

	rawQuery := fmt.Sprintf("UPDATE users SET %s WHERE uuid = :uuid;", strings.Join(fields, ", "))

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	if affected, err := r.db.NamedExecContext(ctx, rawQuery, request); err != nil {
		return err

	} else if !affected {
		return constant.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) UpdateUserPreferred2FAMethod(ctx context.Context, userId, preferred2FAMethod string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/user/UpdateUserPreferred2FAMethod")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `UPDATE users SET preferred_2fa_method = ?, updated_at = ? WHERE uuid = ?`
	affected, err := r.db.ExecContext(ctx, query, preferred2FAMethod, time.Now().UTC(), userId)
	if err != nil {
		r.logger.Error(ctx, "error when updating user preferred 2FA method", logger.Error(err))
		return err
	}

	if !affected {
		return constant.ErrUserNotFound
	}

	return nil
}
