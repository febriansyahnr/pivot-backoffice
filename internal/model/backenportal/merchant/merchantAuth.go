package merchant

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MerchantAuth struct {
	UUID              string         `json:"uuid" db:"uuid"`
	MerchantID        string         `json:"merchantId" db:"merchant_id"`
	Secret            string         `json:"secret" db:"secret"`
	SecretVersion     uint           `json:"-" db:"secret_version"`
	MerchantPublicKey sql.NullString `json:"merchantPublicKey" db:"merchant_public_key"`
	SnapPrivateKey    sql.NullString `json:"snapPrivateKey" db:"snap_private_key"`
	CreatedAt         time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time      `json:"updatedAt" db:"updated_at"`
	DeletedAt         sql.NullTime   `json:"deletedAt" db:"deleted_at"`
}

type MerchantAuthTokenClaims struct {
	jwt.RegisteredClaims
	ClientId   string `json:"clientId"`
	MerchantId string `json:"merchantId"`
}

type AccessTokenB2bRequest struct {
	ClientID     string `json:"clientId" validate:"required"`
	ClientSecret string `json:"clientSecret" validate:"required"`
	GrantType    string `json:"grantType" validate:"required"`
}

type ValidateAccessTokenB2bRequest struct {
	MerchantId    string
	AccessToken   string
	IsSnapRequest bool
}

type SNAPAccessTokenB2BReq struct {
	ClientId  string `json:"-"`
	Timestamp string `json:"-"`
	Signature string `json:"-"`
	GrantType string `json:"grantType"`
}

type SNAPValidateB2b2cTokenSignatureRequest struct {
	ClientId  string `json:"clientId" validate:"required"`
	Timestamp string `json:"timestamp" validate:"required"`
	Signature string `json:"signature" validate:"required"`
}

type AccessTokenB2bResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   string `json:"expiresIn"`
}

type SNAPAccessTokenB2BResp struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	AccessToken     string `json:"accessToken"`
	TokenType       string `json:"tokenType"`
	ExpiresIn       string `json:"expiresIn"`
}

type PKCS8SecretKeyResponse struct {
	MerchantID        string `json:"merchantId"`
	MerchantSecret    string `json:"merchantSecret,omitempty"`
	MerchantPublicKey string `json:"merchantPublicKey,omitempty"`
	SnapPrivateKey    string `json:"snapPrivateKey,omitempty"`
	SnapPublicKey     string `json:"snapPublicKey,omitempty"`
	Data              string `json:"data"` //encrypted data contains all payload below
}

type MerchantPublicKeyRequest struct {
	PublicKey string `json:"publicKey" validate:"required"`
}

type NewMerchantAuthRequest struct {
	ID                  string
	MerchantID          string
	Secret              string
	SecretVersion       uint
	SnapPrivateKey      string
	SnapPrivateKeyValid bool
}

func NewMerchantAuth(req *NewMerchantAuthRequest) *MerchantAuth {
	return &MerchantAuth{
		UUID:          uuid.New().String(),
		MerchantID:    req.MerchantID,
		Secret:        req.Secret,
		SecretVersion: req.SecretVersion,
		SnapPrivateKey: sql.NullString{
			String: req.SnapPrivateKey,
			Valid:  req.SnapPrivateKeyValid,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
