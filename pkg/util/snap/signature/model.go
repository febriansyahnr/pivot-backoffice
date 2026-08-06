package snap_signature

import (
	"encoding/json"
	"fmt"
)

const (
	SignatureTypeTransactional = "transactional"
	SignatureTypeAccessToken   = "access_token"

	PRIVPKCS1 = "RSA PRIVATE KEY"
	PRIVPKCS8 = "PRIVATE KEY"

	PUBPKCS1 = "RSA PUBLIC KEY"
	PUBPKCS8 = "PUBLIC KEY"
)

var (
	errPrivateKeyNotFound = fmt.Errorf("private key not found")
)

type TrxSignature struct {
	HttpMethod                  string
	URL                         string
	AccessToken                 string
	ClientSecret                string
	Timestamp                   string
	BodyPayload                 json.RawMessage
	ClientName                  string
	NeedDoubleHexForString2Sign bool
}
