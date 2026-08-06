package paymentModel

import (
	"github.com/golang-jwt/jwt/v5"
)

type PaymentClaims struct {
	UUID string `json:"uuid"`
	jwt.RegisteredClaims
}
