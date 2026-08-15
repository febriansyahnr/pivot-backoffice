package user

import (
	"encoding/json"
	"time"

	"github.com/skip2/go-qrcode"
)

type EnrollTOTPRequest struct {
	UserId      string `json:"-"`
	QrCodeLevel string `json:"qrCodeLevel" validate:"omitempty,oneof=Low Medium High Highest"`
	QrCodeSize  int    `json:"qrCodeSize"`
}

func (r *EnrollTOTPRequest) GetQrCodeLevel() qrcode.RecoveryLevel {
	switch r.QrCodeLevel {
	default:
		return qrcode.Medium

	case "Low":
		return qrcode.Low

	case "Medium":
		return qrcode.Medium

	case "High":
		return qrcode.High

	case "Highest":
		return qrcode.Highest
	}
}

type ConfirmTOTPRequest struct {
	UserId string `json:"-"`
	OTP    string `json:"otp" validate:"required,numeric,len=6"`
}

type EnrollTOTPResponse struct {
	QRCodeDataURL string `json:"qrCodeDataUrl"`
	SecretKey     string `json:"secretKey"`
}

type ConfirmTOTPResponse struct {
	Status string `json:"status"`
}

type UpdateUserTOTPDataRequest struct {
	UserId         string    `json:"-" db:"uuid"`
	WrappedSecret  string    `json:"-" db:"totp_wrapped_secret"`
	EncryptVersion int       `json:"-" db:"totp_encrypt_version"`
	Status         string    `json:"-" db:"totp_status"`
	UpdatedAt      time.Time `json:"-" db:"updated_at"`
}

type UserTOTPData struct {
	UserId             string    `json:"-" db:"uuid"`
	Email              string    `json:"-" db:"email"`
	TOTPWrappedSecret  *string   `json:"-" db:"totp_wrapped_secret"`
	TOTPEncryptVersion *int      `json:"-" db:"totp_encrypt_version"`
	TOTPStatus         string    `json:"-" db:"totp_status"`
	UpdatedAt          time.Time `json:"-" db:"updated_at"`
}

type TOTPEnrollmentData struct {
	WrappedSecretKey string `json:"wrappedSecretKey" redis:"wrappedSecretKey"`
	EncryptVersion   int    `json:"encryptVersion" redis:"encryptVersion"`
}

func (v TOTPEnrollmentData) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

func (v *TOTPEnrollmentData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, v)
}

type SetPreferred2FAMethodRequest struct {
	UserId             string `json:"-"`
	Preferred2FAMethod string `json:"preferred2FAMethod" validate:"required,oneof=OTP TOTP"`
}

type SetPreferred2FAMethodResponse struct {
	Preferred2FAMethod string `json:"preferred2FAMethod"`
	Updated            bool   `json:"updated"`
}
