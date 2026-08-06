package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestParseRSAPublicKey(t *testing.T) {
	// Generate a test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Convert public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	validPEM := string(pem.EncodeToMemory(pemBlock))

	tests := []struct {
		name    string
		pemStr  string
		wantErr bool
	}{
		{
			name:    "valid RSA public key",
			pemStr:  validPEM,
			wantErr: false,
		},
		{
			name:    "invalid PEM format",
			pemStr:  "invalid pem string",
			wantErr: true,
		},
		{
			name:    "wrong PEM type",
			pemStr:  "-----BEGIN PRIVATE KEY-----\ninvalid\n-----END PRIVATE KEY-----",
			wantErr: true,
		},
		{
			name:    "empty string",
			pemStr:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRSAPublicKey(tt.pemStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRSAPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("ParseRSAPublicKey() returned nil key when no error was expected")
			}
		})
	}
}
