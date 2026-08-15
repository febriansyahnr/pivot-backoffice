package otp

import (
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	jwt "github.com/golang-jwt/jwt/v5"
)

type OTPCache struct {
	OTPList []OTPList `json:"otp" redis:"otp"`
}
type OTPList struct {
	OTP       string    `json:"otp"`
	ExpiredAt time.Time `json:"expired_at"`
	Verify    bool      `json:"verify"`
}

func (o OTPCache) MarshalBinary() ([]byte, error) {
	return json.Marshal(o)
}

func (o *OTPCache) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, o)
}

func (o *OTPCache) GetLatestOTP() *OTPList {
	if len(o.OTPList) == 0 {
		return nil
	}
	return &o.OTPList[len(o.OTPList)-1]
}

type SendOTPReq struct {
	Email               string                   `json:"email" validate:"required,email"`
	TwoFactorAuthMethod constant.TwoFactorMethod `json:"twoFactorAuthMethod" validate:"omitempty,oneof=AUTO OTP TOTP"`
}

type jsonSendOTPReq SendOTPReq

func (s *SendOTPReq) UnmarshalJSON(buf []byte) error {

	data := jsonSendOTPReq{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}

	return nil
}

type VerifyOTPReq struct {
	OTP string `json:"otp" validate:"required,len=6"`
}

type User2FARes struct {
	Token               string                   `json:"token"`
	ResendDelaySeconds  int                      `json:"resendDelaySeconds,omitempty"`
	TwoFactorAuthMethod constant.TwoFactorMethod `json:"twoFactorAuthMethod,omitempty"`
	IsTOTPActive        bool                     `json:"isTOTPActive"`
}

type VerifyOTP struct {
	ID                  string
	Email               string
	OTPCode             string
	TwoFactorAuthMethod constant.TwoFactorMethod
}

type TokenOTPClaims struct {
	UUID      string         `json:"uuid"`
	Email     string         `json:"-"` // Feature authorization
	VerifyOTP VerifyOTPToken `json:"-"` // Verify OTP code
	jwt.RegisteredClaims
}

type SuspendUser struct {
	Status     bool      `json:"status" redis:"status"`
	RetryAfter time.Time `json:"retry_after" redis:"retry_after"`
}

func (s SuspendUser) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *SuspendUser) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, s)
}

type VerifyTOTPRequest struct {
	WrappedSecret  *string
	EncryptVersion *int
	Code           string
}

type GenerateTOTPVerifyTokenRequest struct {
	UserId    string
	UserEmail string
}

type GenerateOTPCodeRequest struct {
	UserId              string
	UserEmail           string
	TwoFactorAuthMethod constant.TwoFactorMethod
}

type VerifyOTPToken struct {
	Token               string                   `json:"token" redis:"token"`
	Email               string                   `json:"email" redis:"email"`
	TwoFactorAuthMethod constant.TwoFactorMethod `json:"twoFactorAuthMethod" redis:"twoFactorAuthMethod"`
}

func (s VerifyOTPToken) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *VerifyOTPToken) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, s)
}
