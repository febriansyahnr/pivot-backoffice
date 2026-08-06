package snap_signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"strings"
)

func NewB2bTokenSignature(s B2bTokenSignature) *B2bTokenSignature {
	return &s
}

type B2bTokenSignature struct {
	ClientID     string
	Timestamp    string
	Signature    string
	ClientSecret string
	Body         interface{}
	PrivateKey   *rsa.PrivateKey
	PublicKey    *rsa.PublicKey
}

func (s *B2bTokenSignature) Create() (string, error) {
	if s.PrivateKey == nil {
		return "", errPrivateKeyNotFound
	}

	stringToSign := s.getApiSign()
	slog.Info("Create B2B Signature string to sign", slog.String("sign", stringToSign))
	chiperBodyHas256 := sha256.New()
	chiperBodyHas256.Write([]byte(stringToSign))
	hashed := chiperBodyHas256.Sum(nil)

	if err := s.PrivateKey.Validate(); err != nil {
		return "", err
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, s.PrivateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	signBase64 := base64.StdEncoding.EncodeToString(signature)

	return signBase64, nil
}

func (s *B2bTokenSignature) getApiSign() string {
	return s.ClientID + "|" + strings.TrimSpace(fmt.Sprint(s.Timestamp))
}

func NewB2bTokenSignatureGenerate(s B2bTokenSignature) (string, error) {
	return s.Create()
}

func (s *B2bTokenSignature) Verify(signature string) bool {
	stringToSign := s.getApiSign() //string2Sign
	slog.Info("Verivy B2B Signature string to sign", slog.String("sign", stringToSign))

	hashed := sha256.Sum256([]byte(stringToSign))

	decodeCipherBodyHashed, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		log.Printf("decode signature error: %v", err)
		return false
	}

	err = rsa.VerifyPKCS1v15(s.PublicKey, crypto.SHA256, hashed[:], decodeCipherBodyHashed)
	if err != nil {
		slog.Info("Verify signature error", " hashed: ", string(hashed[:]), " decodeCipherBodyHashed: ", string(decodeCipherBodyHashed))
		log.Printf(`verify signature error: %v, signature: %s, stringToSign: %s publicKey: %d`, err, signature, stringToSign, s.PublicKey)
		return false
	}

	return true
}
