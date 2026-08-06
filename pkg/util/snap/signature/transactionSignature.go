package snap_signature

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"reflect"
	"strings"
)

func NewTrxSignature(s TrxSignature) (*TrxSignature, error) {
	if s.AccessToken == "" {
		return nil, errors.New("invalid access token")
	}

	token := strings.Split(s.AccessToken, "Bearer ")[1]
	if token == "" {
		return nil, errors.New("invalid access token")
	}

	s = TrxSignature{
		URL:                         strings.TrimSpace(s.URL),
		HttpMethod:                  strings.TrimSpace(strings.ToUpper(s.HttpMethod)),
		AccessToken:                 strings.TrimSpace(token),
		ClientSecret:                strings.TrimSpace(s.ClientSecret),
		Timestamp:                   strings.TrimSpace(s.Timestamp),
		BodyPayload:                 s.BodyPayload,
		NeedDoubleHexForString2Sign: s.NeedDoubleHexForString2Sign,
	}

	return &s, nil
}

func NewTrxSignatureAsymmetric(s TrxSignature) (*TrxSignature, error) {
	s = TrxSignature{
		URL:                         strings.TrimSpace(s.URL),
		HttpMethod:                  strings.TrimSpace(strings.ToUpper(s.HttpMethod)),
		Timestamp:                   strings.TrimSpace(s.Timestamp),
		BodyPayload:                 s.BodyPayload,
		NeedDoubleHexForString2Sign: s.NeedDoubleHexForString2Sign,
	}

	return &s, nil
}

func (s *TrxSignature) Create() *string {
	sha256 := sha256.New()
	sha256.Write(s.BodyPayload)

	sha256SecretKey := strings.ToLower(hex.EncodeToString(sha256.Sum(nil)))

	if s.NeedDoubleHexForString2Sign {
		sha256SecretKey = hex.EncodeToString([]byte(sha256SecretKey))
	}

	hmac512Body := s.HttpMethod + ":" + s.URL + ":" + s.AccessToken + ":" + sha256SecretKey + ":" + s.Timestamp

	slog.Info("Create TrxSignature string to sign", slog.String("client", s.ClientName), slog.String("sign", hmac512Body))

	hmac512 := hmac.New(crypto.SHA512.New, []byte(s.ClientSecret))
	hmac512.Write([]byte(strings.TrimSpace(hmac512Body)))

	signatureToken := base64.StdEncoding.EncodeToString(hmac512.Sum(nil))
	return &signatureToken
}

func (s *TrxSignature) CreateAsymmetric(privateKey *rsa.PrivateKey) (signatureToken string, err error) {
	// we need count param and make to hmac512body

	string2Sign := s.getApiSign()
	slog.Info("Create Asymmetric Signature string to sign", slog.String("sign", string2Sign))

	cipherBodyHash256 := sha256.New()
	cipherBodyHash256.Write([]byte(string2Sign))
	hashed := cipherBodyHash256.Sum(nil)
	if err := privateKey.Validate(); err != nil {
		return "", err
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	sign := base64.StdEncoding.EncodeToString(signature)

	return sign, nil
}

// Verify validates the signature from request.
// It use Symmetric method (HMAC-SHA512)
//
// It takes a signature string and returns a boolean and an error.
func (s *TrxSignature) Verify(signature, clientName string) (bool, error) {
	// validate signature from request
	if s.BodyPayload == nil {
		return false, errors.New("empty body payload")
	}

	if s.URL == "" {
		return false, errors.New("empty url")
	}

	if s.ClientSecret == "" {
		return false, errors.New("empty client secret")
	}

	if s.HttpMethod == "" {
		return false, errors.New("empty http method")
	}

	if s.AccessToken == "" {
		return false, errors.New("empty access token")
	}

	if s.Timestamp == "" {
		return false, errors.New("empty timestamp")
	}

	if clientName != "" {
		s.ClientName = clientName
	}

	signatureToken := s.Create()

	if ok := reflect.DeepEqual(signature, *signatureToken); !ok {
		return false, errors.New("invalid signature")
	}

	return true, nil
}

func (s *TrxSignature) VerifyAsymmetric(publicKey *rsa.PublicKey, signature string) (bool, error) {
	// validate signature from request
	if s.BodyPayload == nil {
		return false, errors.New("empty body payload")
	}

	if s.URL == "" {
		return false, errors.New("empty url")
	}

	if s.Timestamp == "" {
		return false, errors.New("empty timestamp")
	}

	if s.HttpMethod == "" {
		return false, errors.New("empty http method")
	}

	string2Sign := s.getApiSign()

	slog.Info("Verify Asymmetric Signature string to sign", slog.String("sign", string2Sign))
	hashed := sha256.Sum256([]byte(string2Sign))

	decodeCipherBodyHashed, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], decodeCipherBodyHashed)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *TrxSignature) getApiSign() string {
	sha256Body := sha256.New()
	sha256Body.Write(s.BodyPayload)
	bodySign := strings.ToLower(hex.EncodeToString(sha256Body.Sum(nil)))

	if s.NeedDoubleHexForString2Sign {
		bodySign = hex.EncodeToString(sha256Body.Sum(nil))
		bodySign = strings.ToLower(hex.EncodeToString([]byte(bodySign)))
	}

	return s.HttpMethod + ":" + s.URL + ":" + bodySign + ":" + s.Timestamp // signature bnc
}
