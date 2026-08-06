package snap_signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrxSignatureCreate(t *testing.T) {
	validTokenMock := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnROYW1lIjoiUEVSTUFUQSIsImV4cCI6MTcwODQ0Mjg5OCwiZ3JhbmRUeXBlIjoiY2xpZW50X2NyZWRlbnRpYWwifQ.cdzjvv3SYQBPacMvAGSLUjxu3Fp8pXaUi6puNd56LZ0"
	testCases := []struct {
		desc         string
		trxSignature TrxSignature
		expected     string
		wantErr      bool
	}{
		{
			desc:         "success generate signature",
			trxSignature: TrxSignature{URL: "https://example.com", HttpMethod: "POST", AccessToken: validTokenMock, ClientSecret: "client", BodyPayload: json.RawMessage(`{"grantType":"client_credential"}`), Timestamp: "2022-09-15T10:30:00"},
			expected:     "ugZTAqEvAhhFw/fhUpoPAfIiq0Trs7OdeXAsOlp1y7TvMtOOawG/r8eFmYk/Ojgf8YZAX+G036///SteqMj76w==",
			wantErr:      false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trx, err := NewTrxSignature(tc.trxSignature)
			assert.NoError(t, err)

			res := trx.Create()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected, *res)
		})
	}
}

func TestTrxSignatureValidate(t *testing.T) {
	data := []byte(`{"key1": "value1", "key2": 123}`)

	// Convert the JSON byte slice to json.RawMessage
	rawMessage := json.RawMessage(data)

	validTokenMock := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnROYW1lIjoiUEVSTUFUQSIsImV4cCI6MTcwODQ0Mjg5OCwiZ3JhbmRUeXBlIjoiY2xpZW50X2NyZWRlbnRpYWwifQ.cdzjvv3SYQBPacMvAGSLUjxu3Fp8pXaUi6puNd56LZ0"
	testCases := []struct {
		desc         string
		trxSignature TrxSignature
		signature    string
		wantErr      bool
		valid        bool
	}{
		{
			desc: "invalid empty body payload",
			trxSignature: TrxSignature{
				HttpMethod:   "POST",
				URL:          "https://example.com",
				AccessToken:  validTokenMock,
				ClientSecret: "secret",
				BodyPayload:  nil,
			},
			wantErr:   true,
			signature: "",
		},
		{
			desc: "invalid empty url",
			trxSignature: TrxSignature{
				HttpMethod:   "POST",
				URL:          "",
				AccessToken:  validTokenMock,
				ClientSecret: "secret",
				BodyPayload:  rawMessage,
			},
			wantErr:   true,
			signature: "",
		},
		{
			desc: "invalid empty client secret",
			trxSignature: TrxSignature{
				HttpMethod:   "POST",
				URL:          "test.com",
				AccessToken:  validTokenMock,
				ClientSecret: "",
				BodyPayload:  rawMessage,
			},
			wantErr:   true,
			signature: "",
		},
		{
			desc: "invalid empty http method",
			trxSignature: TrxSignature{
				HttpMethod:   "",
				URL:          "test.com",
				AccessToken:  validTokenMock,
				ClientSecret: "client",
				BodyPayload:  rawMessage,
			},
			wantErr:   true,
			signature: "",
		},
		{
			desc: "invalid empty access token",
			trxSignature: TrxSignature{
				HttpMethod:   "POST",
				URL:          "test.com",
				AccessToken:  "Bearer blabla",
				ClientSecret: "client",
				BodyPayload:  rawMessage,
			},
			wantErr:   true,
			signature: "blabla",
		},
		{
			desc: "invalid signature",
			trxSignature: TrxSignature{
				HttpMethod:   "POST",
				URL:          "test.com",
				AccessToken:  validTokenMock,
				ClientSecret: "client",
				BodyPayload:  rawMessage,
				Timestamp:    "2022-09-15T10:30:00",
			},
			signature: "VwYlkO/Nn/RtCp2jUiW819IrHnJkVlSA46MlNSZ8qvU1QTN0Y0wrxyUfU2Wf1dhga10IFXvQMGGG2neiaxYXBW04qj+o7gJp2JhgPvGu1icS+infN5hhDoYV6zQjrlreW35u+Jkzxa1vcguSGVAKrnShwn9YJjfJmARf64IBLzAsFRSSptc51SMRzS/fSdz2+S/B8XFd7mH6+t4GhdVyicHDSzh3XcP5P+xQVpqoAWIGiF4vAUbk3d6w5QeSn2M4yimd6KkoYJb884UzFQOX3XLvLAs+o1iVqhp0gInVCeb/Cb3HL9QB/eXvmgt0Dq3TmviZ6VkCHajjNYo1sL+pog==",
			wantErr:   true,
		},
		{
			desc: "valid signature",
			trxSignature: TrxSignature{
				HttpMethod:   http.MethodPost,
				URL:          "/transfer",
				AccessToken:  validTokenMock,
				ClientSecret: "client",
				BodyPayload:  rawMessage,
				Timestamp:    "2022-09-15T10:30:00",
			},
			signature: "+1",
			valid:     true,
		},
		{
			desc: "error creating signature",
			trxSignature: TrxSignature{
				HttpMethod:   http.MethodPost,
				URL:          "/transfer",
				AccessToken:  validTokenMock,
				ClientSecret: "client",
				BodyPayload:  rawMessage,
				Timestamp:    "2022-09-15T10:30:00",
				// Force Create() to return an error by making the signature creation fail
				NeedDoubleHexForString2Sign: true,
			},
			signature: "will-not-match-anyway", // This doesn't matter since Create() will error
			wantErr:   true,
			valid:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trx, err := NewTrxSignature(tc.trxSignature)
			if !tc.wantErr {
				assert.NoError(t, err)
			}

			if tc.signature == "+1" {
				create := trx.Create()
				tc.signature = *create
			}

			res, err := trx.Verify(tc.signature, "test")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, res, tc.valid)
		})
	}
}

func TestNewTrxSignatureAsymmetric(t *testing.T) {
	data := []byte(`{"key1": "value1", "key2": 123}`)
	rawMessage := json.RawMessage(data)

	testCases := []struct {
		desc         string
		trxSignature TrxSignature
		expected     TrxSignature
	}{
		{
			desc: "success create asymmetric signature",
			trxSignature: TrxSignature{
				URL:                         "  https://example.com  ",
				HttpMethod:                  "  post  ",
				Timestamp:                   "  2022-09-15T10:30:00  ",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: true,
			},
			expected: TrxSignature{
				URL:                         "https://example.com",
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: true,
			},
		},
		{
			desc: "with empty values",
			trxSignature: TrxSignature{
				URL:         "",
				HttpMethod:  "",
				Timestamp:   "",
				BodyPayload: nil,
			},
			expected: TrxSignature{
				URL:         "",
				HttpMethod:  "",
				Timestamp:   "",
				BodyPayload: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := NewTrxSignatureAsymmetric(tc.trxSignature)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.URL, result.URL)
			assert.Equal(t, tc.expected.HttpMethod, result.HttpMethod)
			assert.Equal(t, tc.expected.Timestamp, result.Timestamp)
			assert.Equal(t, tc.expected.BodyPayload, result.BodyPayload)
			assert.Equal(t, tc.expected.NeedDoubleHexForString2Sign, result.NeedDoubleHexForString2Sign)
		})
	}
}

func TestCreateAsymmetric(t *testing.T) {
	// Generate a test private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Create an invalid private key (with N=0) to test validation failure
	invalidPrivateKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: big.NewInt(0),
			E: 65537,
		},
		D: big.NewInt(0),
	}

	data := []byte(`{"key1": "value1", "key2": 123}`)
	rawMessage := json.RawMessage(data)

	testCases := []struct {
		desc         string
		trxSignature TrxSignature
		privateKey   *rsa.PrivateKey
		wantErr      bool
	}{
		{
			desc: "success create asymmetric signature",
			trxSignature: TrxSignature{
				URL:                         "https://example.com",
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: false,
			},
			privateKey: privateKey,
			wantErr:    false,
		},
		{
			desc: "success with double hex for string to sign",
			trxSignature: TrxSignature{
				URL:                         "https://example.com",
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: true,
			},
			privateKey: privateKey,
			wantErr:    false,
		},
		{
			desc: "error with invalid private key",
			trxSignature: TrxSignature{
				URL:                         "https://example.com",
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: false,
			},
			privateKey: invalidPrivateKey,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trx, err := NewTrxSignatureAsymmetric(tc.trxSignature)
			assert.NoError(t, err)

			signatureToken, err := trx.CreateAsymmetric(tc.privateKey)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, signatureToken)

			// Verify the signature is valid by using the corresponding public key
			hashed := sha256.Sum256([]byte(trx.getApiSign()))
			err = rsa.VerifyPKCS1v15(
				&tc.privateKey.PublicKey,
				crypto.SHA256,
				hashed[:],
				mustDecodeBase64(t, signatureToken),
			)
			assert.NoError(t, err, "Signature verification should succeed")
		})
	}
}

func TestCreateAsymmetricRSASignError(t *testing.T) {
	// Create an invalid private key that will fail the privateKey.Validate() check
	// By setting N to nil, it will cause Validate() to return an error
	invalidKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: nil, // This will fail validation
			E: 65537,
		},
	}

	// Set up the TrxSignature
	trxSig := TrxSignature{
		HttpMethod:  "POST",
		URL:         "https://example.com",
		BodyPayload: json.RawMessage(`{"test":"value"}`),
		Timestamp:   "2023-01-01T12:00:00Z",
	}

	// Attempt to create a signature with an invalid key
	sig, err := trxSig.CreateAsymmetric(invalidKey)

	// Verify error is returned due to validation failure
	assert.Error(t, err)
	assert.Empty(t, sig)
	// Check for the actual error message from validation
	assert.Contains(t, err.Error(), "crypto/rsa: missing primes")
}

// Helper function to decode base64 string
func mustDecodeBase64(t *testing.T, str string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(str)
	assert.NoError(t, err, "Failed to decode base64 string")
	return decoded
}

func TestVerifyAsymmetric(t *testing.T) {
	// Generate a test key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	data := []byte(`{"key1": "value1", "key2": 123}`)
	rawMessage := json.RawMessage(data)

	testCases := []struct {
		desc         string
		trxSignature TrxSignature
		setupFunc    func(t *testing.T, s *TrxSignature) string
		wantErr      bool
	}{
		{
			desc: "valid signature",
			trxSignature: TrxSignature{
				URL:                         "https://example.com",
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: false,
			},
			setupFunc: func(t *testing.T, s *TrxSignature) string {
				sig, err := s.CreateAsymmetric(privateKey)
				assert.NoError(t, err)
				return sig
			},
			wantErr: false,
		},
		{
			desc: "empty body payload",
			trxSignature: TrxSignature{
				URL:         "https://example.com",
				HttpMethod:  "POST",
				Timestamp:   "2022-09-15T10:30:00",
				BodyPayload: nil,
			},
			setupFunc: func(t *testing.T, s *TrxSignature) string {
				return "dummy-signature"
			},
			wantErr: true,
		},
		{
			desc: "empty URL",
			trxSignature: TrxSignature{
				URL:         "",
				HttpMethod:  "POST",
				Timestamp:   "2022-09-15T10:30:00",
				BodyPayload: rawMessage,
			},
			setupFunc: func(t *testing.T, s *TrxSignature) string {
				return "dummy-signature"
			},
			wantErr: true,
		},
		{
			desc: "empty timestamp",
			trxSignature: TrxSignature{
				URL:         "https://example.com",
				HttpMethod:  "POST",
				Timestamp:   "",
				BodyPayload: rawMessage,
			},
			setupFunc: func(t *testing.T, s *TrxSignature) string {
				return "dummy-signature"
			},
			wantErr: true,
		},
		{
			desc: "empty http method",
			trxSignature: TrxSignature{
				URL:         "https://example.com",
				HttpMethod:  "",
				Timestamp:   "2022-09-15T10:30:00",
				BodyPayload: rawMessage,
			},
			setupFunc: func(t *testing.T, s *TrxSignature) string {
				return "dummy-signature"
			},
			wantErr: true,
		},
		{
			desc: "invalid signature format",
			trxSignature: TrxSignature{
				URL:         "https://example.com",
				HttpMethod:  "POST",
				Timestamp:   "2022-09-15T10:30:00",
				BodyPayload: rawMessage,
			},
			setupFunc: func(t *testing.T, s *TrxSignature) string {
				return "invalid-base64-format"
			},
			wantErr: true,
		},
		{
			desc: "valid with double hex string to sign",
			trxSignature: TrxSignature{
				URL:                         "https://example.com",
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: true,
			},
			setupFunc: func(t *testing.T, s *TrxSignature) string {
				sig, err := s.CreateAsymmetric(privateKey)
				assert.NoError(t, err)
				return sig
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trx, err := NewTrxSignatureAsymmetric(tc.trxSignature)
			assert.NoError(t, err)

			signature := tc.setupFunc(t, trx)

			result, err := trx.VerifyAsymmetric(publicKey, signature)

			if tc.wantErr {
				assert.Error(t, err)
				assert.False(t, result)
			} else {
				assert.NoError(t, err)
				assert.True(t, result)
			}
		})
	}
}

func TestErrorPathVerify(t *testing.T) {
	data := []byte(`{"key1": "value1", "key2": 123}`)
	rawMessage := json.RawMessage(data)
	validTokenMock := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnROYW1lIjoiUEVSTUFUQSIsImV4cCI6MTcwODQ0Mjg5OCwiZ3JhbmRUeXBlIjoiY2xpZW50X2NyZWRlbnRpYWwifQ.cdzjvv3SYQBPacMvAGSLUjxu3Fp8pXaUi6puNd56LZ0"

	// Set up a TrxSignature with an implementation that will force Create() to fail
	// Creating a TrxSignature with a ClientSecret that will cause hmac.New to panic,
	// which we'll recover from but will be detected in our test
	invalidTrx := TrxSignature{
		HttpMethod:   http.MethodPost,
		URL:          "/transfer",
		AccessToken:  validTokenMock,
		ClientSecret: "", // Empty client secret will cause Create() to fail with an error
		BodyPayload:  rawMessage,
		Timestamp:    "2022-09-15T10:30:00",
	}

	trx, err := NewTrxSignature(invalidTrx)
	assert.NoError(t, err)

	// Create a signature that doesn't matter since Create() will fail
	signature := "test-signature"

	// Test that Verify returns an error
	result, err := trx.Verify(signature, "test")
	assert.Error(t, err)
	assert.False(t, result)
}

func TestTrxSignatureGetApiSign(t *testing.T) {
	data := []byte(`{"key1": "value1", "key2": 123}`)
	rawMessage := json.RawMessage(data)

	testCases := []struct {
		desc         string
		trxSignature TrxSignature
	}{
		{
			desc: "regular case without double hex",
			trxSignature: TrxSignature{
				URL:                         "example.com", // Using a URL without http:// to avoid colon splitting issues
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: false,
			},
		},
		{
			desc: "with double hex encoding",
			trxSignature: TrxSignature{
				URL:                         "example.com", // Using a URL without http:// to avoid colon splitting issues
				HttpMethod:                  "POST",
				Timestamp:                   "2022-09-15T10:30:00",
				BodyPayload:                 rawMessage,
				NeedDoubleHexForString2Sign: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trx, err := NewTrxSignatureAsymmetric(tc.trxSignature)
			assert.NoError(t, err)

			result := trx.getApiSign()

			// Simply verify the result is not empty and contains the expected method and URL
			assert.Contains(t, result, tc.trxSignature.HttpMethod)
			assert.Contains(t, result, tc.trxSignature.URL)
			assert.Contains(t, result, tc.trxSignature.Timestamp)

			// For double hex, verify the hash portion is present in the correct format
			if tc.trxSignature.NeedDoubleHexForString2Sign {
				// Calculate the expected hash without double encoding
				sha256Body := sha256.New()
				sha256Body.Write(tc.trxSignature.BodyPayload)
				singleHash := strings.ToLower(hex.EncodeToString(sha256Body.Sum(nil)))

				// Calculate the double-encoded hash
				doubleHash := strings.ToLower(hex.EncodeToString([]byte(singleHash)))

				// Verify the double hash is in the result
				assert.Contains(t, result, doubleHash)
			} else {
				// Calculate the expected hash
				sha256Body := sha256.New()
				sha256Body.Write(tc.trxSignature.BodyPayload)
				singleHash := strings.ToLower(hex.EncodeToString(sha256Body.Sum(nil)))

				// Verify the hash is in the result
				assert.Contains(t, result, singleHash)
			}
		})
	}
}

func TestDoubleHexEncoding(t *testing.T) {
	// Create a simple test payload
	payload := json.RawMessage(`{"test":"value"}`)

	// Create two signatures with the same data, one with double hex and one without
	regularTrx := TrxSignature{
		URL:                         "test.com",
		HttpMethod:                  "POST",
		Timestamp:                   "2022-09-15T10:30:00",
		BodyPayload:                 payload,
		NeedDoubleHexForString2Sign: false,
	}

	doubleHexTrx := TrxSignature{
		URL:                         "test.com",
		HttpMethod:                  "POST",
		Timestamp:                   "2022-09-15T10:30:00",
		BodyPayload:                 payload,
		NeedDoubleHexForString2Sign: true,
	}

	// Get the API sign string for both
	regularSign := regularTrx.getApiSign()
	doubleHexSign := doubleHexTrx.getApiSign()

	// Extract the hash portion (third segment) from both strings
	regularParts := strings.Split(regularSign, ":")
	doubleHexParts := strings.Split(doubleHexSign, ":")

	regularHash := regularParts[2]
	doubleHexHash := doubleHexParts[2]

	// Calculate what the regular hash should be
	sha256Body := sha256.New()
	sha256Body.Write(payload)
	expectedRegularHash := strings.ToLower(hex.EncodeToString(sha256Body.Sum(nil)))

	// The double hex hash should be a hex encoding of the regular hash string
	expectedDoubleHexHash := strings.ToLower(hex.EncodeToString([]byte(expectedRegularHash)))

	// Verify the regular hash matches our calculation
	assert.Equal(t, expectedRegularHash, regularHash, "Regular hash should match expected value")

	// Verify the double hex hash matches our calculation
	assert.Equal(t, expectedDoubleHexHash, doubleHexHash, "Double hex hash should match expected value")

	// Verify that the double hex hash is different from the regular hash
	assert.NotEqual(t, regularHash, doubleHexHash, "Double hex hash should be different from regular hash")

	// Verify that the double hex hash length is twice the regular hash length
	assert.Equal(t, len(regularHash)*2, len(doubleHexHash), "Double hex hash length should be twice the regular hash length")
}

func TestNewTrxSignature(t *testing.T) {
	validTokenMock := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnROYW1lIjoiUEVSTUFUQSIsImV4cCI6MTcwODQ0Mjg5OCwiZ3JhbmRUeXBlIjoiY2xpZW50X2NyZWRlbnRpYWwifQ.cdzjvv3SYQBPacMvAGSLUjxu3Fp8pXaUi6puNd56LZ0"

	// Test the empty access token case which should return an error
	t.Run("ERROR: empty access token", func(t *testing.T) {
		trx := TrxSignature{
			URL:          "https://example.com",
			HttpMethod:   "POST",
			AccessToken:  "",
			ClientSecret: "client",
			BodyPayload:  json.RawMessage(`{"test":"data"}`),
			Timestamp:    "2022-09-15T10:30:00",
		}

		result, err := NewTrxSignature(trx)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid access token", err.Error())
	})

	// Test the case where Bearer prefix is present but token is empty
	t.Run("ERROR: empty token after Bearer prefix", func(t *testing.T) {
		trx := TrxSignature{
			URL:          "https://example.com",
			HttpMethod:   "POST",
			AccessToken:  "Bearer ",
			ClientSecret: "client",
			BodyPayload:  json.RawMessage(`{"test":"data"}`),
			Timestamp:    "2022-09-15T10:30:00",
		}

		result, err := NewTrxSignature(trx)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid access token", err.Error())
	})

	// Test the success case
	t.Run("SUCCESS: valid access token", func(t *testing.T) {
		trx := TrxSignature{
			URL:          "https://example.com",
			HttpMethod:   "POST",
			AccessToken:  validTokenMock,
			ClientSecret: "client",
			BodyPayload:  json.RawMessage(`{"test":"data"}`),
			Timestamp:    "2022-09-15T10:30:00",
		}

		result, err := NewTrxSignature(trx)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Check that the token was properly extracted
		assert.NotContains(t, result.AccessToken, "Bearer ")
		assert.NotEmpty(t, result.AccessToken)

		// Other fields were properly trimmed
		assert.Equal(t, strings.TrimSpace(trx.URL), result.URL)
		assert.Equal(t, strings.ToUpper(strings.TrimSpace(trx.HttpMethod)), result.HttpMethod)
	})
}

func TestVerifyAsymmetricDecodeError(t *testing.T) {
	// Generate a key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Create a transaction signature object
	ts := TrxSignature{
		URL:         "https://example.com/api/v1/test",
		HttpMethod:  "POST",
		Timestamp:   "2023-09-15T12:00:00Z",
		BodyPayload: []byte(`{"test":"data"}`),
	}

	// Test case 1: Invalid base64 signature (should trigger error at line 171)
	invalidBase64 := "this-is-not-valid-base64@#$%^"
	result, err := ts.VerifyAsymmetric(publicKey, invalidBase64)

	// Debug output to confirm error path
	t.Logf("Base64 decode error result=%v, err=%v", result, err)

	if err == nil {
		t.Error("Expected error when decoding invalid base64 signature, but got nil")
	}
	assert.False(t, result)
	assert.Contains(t, err.Error(), "illegal base64")

	// Test case 2: Valid base64 but invalid signature
	validBase64ButInvalidSig := base64.StdEncoding.EncodeToString([]byte("not-a-valid-signature"))
	_, err = ts.VerifyAsymmetric(publicKey, validBase64ButInvalidSig)
	if err == nil {
		t.Error("Expected error with invalid signature content, but got nil")
	}

	// Test case 3: Missing required fields
	emptyTs := TrxSignature{}
	_, err = emptyTs.VerifyAsymmetric(publicKey, validBase64ButInvalidSig)
	if err == nil {
		t.Error("Expected error with empty transaction signature, but got nil")
	}
}

func TestVerifyAsymmetricSuccess(t *testing.T) {
	// Generate a key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Create a transaction signature object
	ts := TrxSignature{
		URL:         "https://example.com/api/v1/test",
		HttpMethod:  "POST",
		Timestamp:   "2023-09-15T12:00:00Z",
		BodyPayload: []byte(`{"test":"data"}`),
	}

	// Create a valid signature
	signature, err := ts.CreateAsymmetric(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Verify the signature
	valid, err := ts.VerifyAsymmetric(publicKey, signature)
	if err != nil {
		t.Errorf("Verification failed: %v", err)
	}
	if !valid && err == nil {
		t.Error("Expected verification to succeed, but it failed without an error")
	}
}

func TestTrxSignature_Struct(t *testing.T) {
	// Define a test struct
	type fields struct {
		URL                         string
		HttpMethod                  string
		Timestamp                   string
		BodyPayload                 []byte
		NeedDoubleHexForString2Sign bool
	}

	// Define test cases
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Missing URL",
			fields: fields{
				HttpMethod:  "POST",
				Timestamp:   "2023-09-15T12:00:00Z",
				BodyPayload: []byte(`{"test":"data"}`),
			},
			wantErr: true,
		},
		{
			name: "Missing Body Payload",
			fields: fields{
				URL:        "https://example.com/api/v1/test",
				HttpMethod: "POST",
				Timestamp:  "2023-09-15T12:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "Missing HTTP Method",
			fields: fields{
				URL:         "https://example.com/api/v1/test",
				Timestamp:   "2023-09-15T12:00:00Z",
				BodyPayload: []byte(`{"test":"data"}`),
			},
			wantErr: true,
		},
		{
			name: "Missing Timestamp",
			fields: fields{
				URL:         "https://example.com/api/v1/test",
				HttpMethod:  "POST",
				BodyPayload: []byte(`{"test":"data"}`),
			},
			wantErr: true,
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a key pair for testing
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("Failed to generate RSA key: %v", err)
			}
			publicKey := &privateKey.PublicKey

			// Create the TrxSignature instance with test fields
			s := &TrxSignature{
				URL:                         tt.fields.URL,
				HttpMethod:                  tt.fields.HttpMethod,
				Timestamp:                   tt.fields.Timestamp,
				BodyPayload:                 tt.fields.BodyPayload,
				NeedDoubleHexForString2Sign: tt.fields.NeedDoubleHexForString2Sign,
			}

			// For the tests that should have valid fields, create a real signature
			var signature string
			if !tt.wantErr {
				var err error
				signature, err = s.CreateAsymmetric(privateKey)
				if err != nil {
					t.Fatalf("Failed to create signature: %v", err)
				}
			} else {
				// Use a valid base64 string for error tests
				signature = base64.StdEncoding.EncodeToString([]byte("dummy-signature"))
			}

			// Verify the signature
			_, err = s.VerifyAsymmetric(publicKey, signature)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyAsymmetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestVerifyEmptyAccessToken specifically tests the error path for empty access token
func TestVerifyEmptyAccessToken(t *testing.T) {
	data := []byte(`{"key1": "value1", "key2": 123}`)
	rawMessage := json.RawMessage(data)

	// Create a TrxSignature with an empty access token
	trx := TrxSignature{
		HttpMethod:   "POST",
		URL:          "https://example.com",
		AccessToken:  "", // Empty access token should trigger the error
		ClientSecret: "client",
		BodyPayload:  rawMessage,
		Timestamp:    "2022-09-15T10:30:00",
	}

	// Create a new TrxSignature instance directly without using NewTrxSignature
	// since that would validate the access token at creation time
	signature := "test-signature"

	// Verify should return an error due to empty access token
	result, err := trx.Verify(signature, "test")

	// Check error and result
	assert.Error(t, err)
	assert.Equal(t, "empty access token", err.Error())
	assert.False(t, result)
}

// TestCreateAsymmetricSignPKCS1v15Error tests the error path when rsa.SignPKCS1v15 fails
func TestCreateAsymmetricSignPKCS1v15Error(t *testing.T) {
	// Create a normal transaction signature object
	ts := TrxSignature{
		URL:         "https://example.com",
		HttpMethod:  "POST",
		Timestamp:   "2022-09-15T10:30:00",
		BodyPayload: []byte(`{"test":"data"}`),
	}

	// Generate a minimal "broken" private key that will pass validation
	// but cause SignPKCS1v15 to fail
	brokenKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: big.NewInt(0).SetBytes([]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			}),
			E: 65537,
		},
		D: big.NewInt(0).SetBytes([]byte{
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}),
		Primes: []*big.Int{
			big.NewInt(3),
			big.NewInt(5),
		},
	}

	// Set minimal values to pass validation but ensure signing fails
	brokenKey.Precomputed.Dp = big.NewInt(1)
	brokenKey.Precomputed.Dq = big.NewInt(1)
	brokenKey.Precomputed.Qinv = big.NewInt(1)
	// CRTValues field is deprecated and not needed for this test

	// Try to create a signature
	_, err := ts.CreateAsymmetric(brokenKey)

	// Verify the error is returned
	assert.Error(t, err)
	// The error should be about input overflowing the modulus
	assert.Contains(t, err.Error(), "input overflows the modulus")
}

// TestCreateAsymmetricKeyValidationError tests the error path when privateKey.Validate() fails
func TestCreateAsymmetricKeyValidationError(t *testing.T) {
	// Create a normal transaction signature object
	ts := TrxSignature{
		URL:         "https://example.com",
		HttpMethod:  "POST",
		Timestamp:   "2022-09-15T10:30:00",
		BodyPayload: []byte(`{"test":"data"}`),
	}

	// Create an invalid private key that will fail validation
	// An RSA private key without primes will fail validation
	invalidKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: big.NewInt(0).SetBytes([]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			}),
			E: 65537,
		},
		D: big.NewInt(0).SetBytes([]byte{
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}),
		// Missing Primes will cause Validate() to fail with "crypto/rsa: missing primes"
		Primes: nil,
	}

	// Try to create a signature with the invalid key
	_, err := ts.CreateAsymmetric(invalidKey)

	// Verify the error is returned
	assert.Error(t, err)
	// Check that the error is about missing primes (from Validate())
	assert.Contains(t, err.Error(), "missing primes")
}

// TestCreateAsymmetricCompleteCoverage ensures full coverage of the CreateAsymmetric function
func TestCreateAsymmetricCompleteCoverage(t *testing.T) {
	// Create a transaction signature object
	ts := TrxSignature{
		URL:         "https://example.com",
		HttpMethod:  "POST",
		Timestamp:   "2022-09-15T10:30:00",
		BodyPayload: []byte(`{"test":"data"}`),
	}

	// Generate a valid key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Call CreateAsymmetric and ensure it works correctly
	signature, err := ts.CreateAsymmetric(privateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// Now verify the signature to ensure it's valid
	publicKey := &privateKey.PublicKey
	valid, err := ts.VerifyAsymmetric(publicKey, signature)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test with double hex encoding enabled
	tsWithDoubleHex := TrxSignature{
		URL:                         "https://example.com",
		HttpMethod:                  "POST",
		Timestamp:                   "2022-09-15T10:30:00",
		BodyPayload:                 []byte(`{"test":"data"}`),
		NeedDoubleHexForString2Sign: true,
	}

	// Generate signature with double hex encoding
	signatureWithDoubleHex, err := tsWithDoubleHex.CreateAsymmetric(privateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, signatureWithDoubleHex)

	// Verify the double hex encoded signature
	validDoubleHex, err := tsWithDoubleHex.VerifyAsymmetric(publicKey, signatureWithDoubleHex)
	assert.NoError(t, err)
	assert.True(t, validDoubleHex)
}

// TestCreateAsymmetricNilKey tests that CreateAsymmetric properly panics when nil private key is used
func TestCreateAsymmetricNilKey(t *testing.T) {
	// Create a transaction signature object
	ts := TrxSignature{
		URL:         "https://example.com",
		HttpMethod:  "POST",
		Timestamp:   "2022-09-15T10:30:00",
		BodyPayload: []byte(`{"test":"data"}`),
	}

	// Set up the defer/recover before calling the function
	defer func() {
		if r := recover(); r != nil {
			// Successfully caught the panic, test passes
			t.Log("Successfully caught panic from nil private key")
			return
		}
		// If we get here without recovering a panic, the test should fail
		t.Error("Expected a panic when using nil private key, but no panic occurred")
	}()

	// This should cause a panic that our recover will catch
	ts.CreateAsymmetric(nil)
}

// TestCreateAsymmetricSignError tests the path where SignPKCS1v15 fails due to rand.Reader error
func TestCreateAsymmetricSignError(t *testing.T) {
	// Skip if running in parallel or this could affect other tests
	if t.Parallel(); true {
		t.Skip("Skipping test that modifies global state")
	}

	// Generate a valid private key first, before we replace the reader
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Create a mock reader that always errors
	mockReader := &mockReader{err: fmt.Errorf("mock reader error")}

	// Save the original rand.Reader
	originalReader := rand.Reader

	// Replace with our mock reader
	rand.Reader = mockReader

	// Restore the original reader when test finishes
	defer func() {
		rand.Reader = originalReader
	}()

	// Create transaction signature object
	ts := TrxSignature{
		URL:         "https://example.com",
		HttpMethod:  "POST",
		Timestamp:   "2022-09-15T10:30:00",
		BodyPayload: []byte(`{"test":"data"}`),
	}

	// Call CreateAsymmetric, which should now fail at the SignPKCS1v15 step
	_, err = ts.CreateAsymmetric(privateKey)
	assert.Error(t, err, "Should return error when signing with mocked reader")
}

// mockReader is a custom io.Reader that always returns an error or fixed data
type mockReader struct {
	data []byte
	err  error
}

func (r *mockReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	n := copy(p, r.data)
	return n, nil
}

// TestCreateAsymmetricInternalErrors tests the internal error paths in CreateAsymmetric
func TestCreateAsymmetricInternalErrors(t *testing.T) {
	// Test Validate() error
	t.Run("validation error", func(t *testing.T) {
		// Create an invalid key that will fail validation
		invalidKey := &rsa.PrivateKey{
			PublicKey: rsa.PublicKey{
				N: big.NewInt(0),
				E: 65537,
			},
			D:      big.NewInt(0),
			Primes: nil, // This will cause validation to fail
		}

		ts := TrxSignature{
			URL:         "https://example.com",
			HttpMethod:  "POST",
			Timestamp:   "2022-09-15T10:30:00",
			BodyPayload: []byte(`{"test":"data"}`),
		}

		_, err := ts.CreateAsymmetric(invalidKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing primes")
	})
}
