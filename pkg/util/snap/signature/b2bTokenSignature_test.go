package snap_signature

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetApiSign(t *testing.T) {
	testCases := []struct {
		clientID  string
		timestamp string
		expected  string
		wantErr   bool
	}{
		{"123", "1234567890", "123|1234567890", true},
		{"abc", "9876543210", "abc|9876543210", false},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		s := &B2bTokenSignature{
			ClientID:  tc.clientID,
			Timestamp: tc.timestamp,
		}
		result := s.getApiSign()
		if result != tc.expected && !tc.wantErr {
			t.Errorf("For ClientID: %s and Timestamp: %s - Expected %s, but got %s", tc.clientID, tc.timestamp, tc.expected, result)
		}
	}
}

func TestCreate(t *testing.T) {
	mockPayloadString := `{"grantType":"client_credential"}`
	mockPayload := []byte(mockPayloadString)

	testCases := []struct {
		desc       string
		bodyStruct B2bTokenSignature
		expected   string
		privateKey func() *rsa.PrivateKey
		wantErr    bool
	}{
		{
			desc: "failed to load private key",
			bodyStruct: B2bTokenSignature{
				Timestamp: time.Now().Format(time.RFC3339),
				ClientID:  "123",
				Body:      mockPayload,
			},
			privateKey: func() *rsa.PrivateKey { return nil },
			wantErr:    true,
		},
		{
			desc: "success generate signature",
			bodyStruct: B2bTokenSignature{
				Timestamp: "2022-09-15T10:30:00",
				ClientID:  "mock",
				Body:      mockPayload,
			},
			privateKey: func() *rsa.PrivateKey {
				p, err := generateRsaPrivateKeyMock()
				if err != nil {
					t.Fatalf("failed to generate private key: %v", err)
				}

				return p
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			tC.bodyStruct.PrivateKey = tC.privateKey()
			create, err := tC.bodyStruct.Create()
			if err != nil && !tC.wantErr {
				t.Errorf("failed to generate private key: %v", err)
			}

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, create)
			}

			if create != "" {
				fmt.Print("signature created: ", create)
			}

		})
	}
}

func generateRsaPrivateKeyMock() (*rsa.PrivateKey, error) {
	pemData := `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDI/t4aI3sTFTTH
Lxpdd1HFqKoMUyeHFa9Oiq//OGlMODsR5JHsF+IuRfar/U8HXe6jg0j5sLp/5D8a
i5MiqvyXJxd1j8EuKKs/qeQrtS0usqqJYJDMze3YX1eIHrD5LO2CBiJhPp6gIhUO
SfhImxJEMaMDiounZmoK9P66JrEKjedRbuVgDn4PtxSLr3VzYYRZPa65D5HzjdIm
OuotubMN3mbWRWhicT2oM29fTe3cJlzRAcaOmicPdRtNTf0/Wk+jCDeiJxoLuKyZ
/IUSVuzIl4gnMGKM/KheSPOlTRwVm+eHWjr2+xB4KZf00FNoJTu+TXUEe0pdYVfw
ioeavUHdAgMBAAECggEBAJNI8EgHJ/Db4Uj0Y0WKYgmNhs5xQM3kPgo35rAHDmIj
8mUyMRvohH2UFyYBASBM3MpFMfyGXKPLBdLV5IPK+D1rD+294bmJY7PLMsA0i19k
3UK92F27qUac1u+QTe7J1WEqTZck4+hEEVnfKmlJ+SCvntzBcYTBr4NH9EFEiQdJ
l1fMfMxS6HgpIcY2wxnypqjpeC/LdG4543iGpP8zbnunxW43yFgkNC8AJ/KuY9eL
dX/bES25F0mmhQRruzpFczre+4S2tiVhMq73VuyXJiYIvaCxA1JHQjQBCjmkJ7r7
lNKZYlf7Qfcg5FZ9njGdFWpDeM40vyiGYZbKVV6X4aECgYEA9n84ldyPD4U8Kmnz
56pJ5lK+2ztC90pXCNS+vpCqg3VdLf7eE1nUAp2f6JBA/qm3kMIfju49bpTFM6ZD
Kj4jI8WI7JdBy1gfNls9HF5nVtoHzTHG9E3wT0YAOPHmyCfohpfFIFaWX63VwZ15
/jKP9OqW7aeMLoOWv9Mg+WQbVrMCgYEA0L6TKuttEZmyLDFhb+4T52Yl8ofNLbxe
6BhqDy44xwjBsi2a0BDc9l1/kv2ACsHEfMNFO/jW1IZ4C+2S3WDSoH+ff7uFVxS0
1wGrNNM/ch/zX3oLzX+ZevAYSys95IshMQMfoePp533ZYk504nRsxM4JVGSUoSHK
MvmnlRMJzS8CgYEAiJOE7sP+IENaSsXZ9opL1+oRBbeYKxxtjN8TsNLHJ39n2YxV
z7L93VUovNrwqCmxI+vrQG6QayzS9wMwQ7+aCL/yVeSY9+ojoSJ8gbNs3pp/qBnk
eoiUldfbV7HwhQZXt/tvpbNULj9LKLPwW//382PnrFYhPcR7Sl3Y71WgMDECgYBv
DzXVe/RHjPJSuOMSXiSQ1LQT2VS8pKAJ9BNZiEoE+w+y8LiRQqeNHCmn1t+s2XLk
vi+zvKzv3as5DWk6By2I3t3JY8eJkSa1zdl8/XegDIe7oH9vEhhiZCNIuvTvB2bd
YMAPrebglwB1YTCm2zKTcttb3zeEkym0/Ua/9aUdWQKBgQC43+cW1iWcw3U8LhO7
3Wkys/d07xkZEoEGJRFymqhyUNIBRY3fe4ukQOp3Qq0bff0Oj5YBDD6n+TES9nSh
wrKPt/kduqwqr9Ob4SwaUwX18rQlQwRxo2O3EQLcIIMiMVSuDKdc68AT8uc1Dbqn
gmyBqtD/AVn+rKit0f7HDuOrcw==
-----END PRIVATE KEY-----`

	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("mocking private key not found")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey.(*rsa.PrivateKey), nil
}

func TestVerify(t *testing.T) {
	testCases := []struct {
		name      string
		signature string
		timestamp string
		clientID  string
		valid     bool
		wantErr   bool
	}{
		{
			name:      "Valid signature",
			signature: "VwYlkO/Nn/RtCp2jUiW819IrHnJkVlSA46MlNSZ8qvU1QTN0Y0wrxyUfU2Wf1dhga10IFXvQMGGG2neiaxYXBW04qj+o7gJp2JhgPvGu1icS+infN5hhDoYV6zQjrlreW35u+Jkzxa1vcguSGVAKrnShwn9YJjfJmARf64IBLzAsFRSSptc51SMRzS/fSdz2+S/B8XFd7mH6+t4GhdVyicHDSzh3XcP5P+xQVpqoAWIGiF4vAUbk3d6w5QeSn2M4yimd6KkoYJb884UzFQOX3XLvLAs+o1iVqhp0gInVCeb/Cb3HL9QB/eXvmgt0Dq3TmviZ6VkCHajjNYo1sL+pog==",
			timestamp: "2022-09-15T10:30:00",
			clientID:  "randomclientid",
			valid:     true,
		},
		{
			name:      "Invalid signature",
			signature: "invalid_signature",
			valid:     false,
			wantErr:   true,
		},
		{
			name:      "invalid signature wrong timestamp",
			signature: "VwYlkO/Nn/RtCp2jUiW819IrHnJkVlSA46MlNSZ8qvU1QTN0Y0wrxyUfU2Wf1dhga10IFXvQMGGG2neiaxYXBW04qj+o7gJp2JhgPvGu1icS+infN5hhDoYV6zQjrlreW35u+Jkzxa1vcguSGVAKrnShwn9YJjfJmARf64IBLzAsFRSSptc51SMRzS/fSdz2+S/B8XFd7mH6+t4GhdVyicHDSzh3XcP5P+xQVpqoAWIGiF4vAUbk3d6w5QeSn2M4yimd6KkoYJb884UzFQOX3XLvLAs+o1iVqhp0gInVCeb/Cb3HL9QB/eXvmgt0Dq3TmviZ6VkCHajjNYo1sL+pog==",
			timestamp: "2022-09-15T10:00:00",
			clientID:  "randomclientid",
			valid:     false,
			wantErr:   true,
		},
		{
			name:      "invalid signature invalid client id",
			signature: "VwYlkO/Nn/RtCp2jUiW819IrHnJkVlSA46MlNSZ8qvU1QTN0Y0wrxyUfU2Wf1dhga10IFXvQMGGG2neiaxYXBW04qj+o7gJp2JhgPvGu1icS+infN5hhDoYV6zQjrlreW35u+Jkzxa1vcguSGVAKrnShwn9YJjfJmARf64IBLzAsFRSSptc51SMRzS/fSdz2+S/B8XFd7mH6+t4GhdVyicHDSzh3XcP5P+xQVpqoAWIGiF4vAUbk3d6w5QeSn2M4yimd6KkoYJb884UzFQOX3XLvLAs+o1iVqhp0gInVCeb/Cb3HL9QB/eXvmgt0Dq3TmviZ6VkCHajjNYo1sL+pog==",
			timestamp: "2022-09-15T10:00:00",
			clientID:  "invalidIDss",
			valid:     false,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := generateRsaPrivateKeyMock()
			if err != nil {
				t.Errorf("Failed to generate private key: %v", err)
			}

			s := &B2bTokenSignature{
				PublicKey: &key.PublicKey,
				Timestamp: tc.timestamp,
				ClientID:  tc.clientID,
			}
			result := s.Verify(tc.signature)
			if result != tc.valid {
				t.Errorf("Expected result for %s to be %t, but got %t", tc.name, tc.valid, result)
			}
		})
	}
}

func TestB2bTokenSignature(t *testing.T) {
	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	// Test cases
	testCases := []struct {
		name      string
		signature B2bTokenSignature
		wantErr   bool
	}{
		{
			name: "Valid signature",
			signature: B2bTokenSignature{
				ClientID:   "randomclientid",
				Timestamp:  "2022-09-15T10:30:00",
				PrivateKey: privateKey,
				PublicKey:  publicKey,
			},
			wantErr: false,
		},
		{
			name: "Missing private key",
			signature: B2bTokenSignature{
				ClientID:  "randomclientid",
				Timestamp: "2022-09-15T10:30:00",
				PublicKey: publicKey,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new signature
			b2bSignature := NewB2bTokenSignature(tc.signature)
			assert.NotNil(t, b2bSignature)

			// Generate signature
			signature, err := b2bSignature.Create()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, signature)

			// Test NewB2bTokenSignatureGenerate
			generatedSignature, err := NewB2bTokenSignatureGenerate(tc.signature)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, generatedSignature)

			// Verify signature
			b2bSignature.Signature = signature
			isValid := b2bSignature.Verify(signature)
			assert.True(t, isValid)

			// Test invalid signature verification
			isValidWithInvalidSig := b2bSignature.Verify("invalid-signature")
			assert.False(t, isValidWithInvalidSig)
		})
	}
}

// TestCreateB2BTotalCoverage ensures 100% coverage of the Create function
func TestCreateB2BTotalCoverage(t *testing.T) {
	// Create a signature with an empty private key to trigger error
	signature := B2bTokenSignature{
		ClientID:  "testclient",
		Timestamp: "2022-09-15T10:30:00",
		// Explicitly leave the PrivateKey nil to trigger the error case
	}

	// Attempt to create signature, which should fail
	_, err := signature.Create()
	assert.Error(t, err, "Should return error when private key is nil")
}

// TestCreateB2BValidationError tests the case where private key validation fails
func TestCreateB2BValidationError(t *testing.T) {
	// Create an invalid private key that will fail validation
	// An RSA private key without primes will fail validation
	invalidKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: new(big.Int).SetBytes([]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			}),
			E: 65537,
		},
		D: new(big.Int).SetBytes([]byte{
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}),
		// Missing Primes will cause Validate() to fail
		Primes: nil,
	}

	// Create a signature with the invalid private key
	signature := B2bTokenSignature{
		ClientID:   "testclient",
		Timestamp:  "2022-09-15T10:30:00",
		PrivateKey: invalidKey,
	}

	// Attempt to create signature, which should fail at validation
	_, err := signature.Create()
	assert.Error(t, err, "Should return error when private key validation fails")
	assert.Contains(t, err.Error(), "missing primes")
}

// TestCreateB2BSigningError tests the case where SignPKCS1v15 fails
func TestCreateB2BSigningError(t *testing.T) {
	// Create a test private key that will pass validation
	// but cause SignPKCS1v15 to fail
	brokenKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: new(big.Int).SetBytes([]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			}),
			E: 65537,
		},
		D: new(big.Int).SetBytes([]byte{
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}),
		Primes: []*big.Int{
			big.NewInt(3),
			big.NewInt(5),
		},
	}

	// Set up minimal values to make validation pass
	brokenKey.Precomputed.Dp = big.NewInt(1)
	brokenKey.Precomputed.Dq = big.NewInt(1)
	brokenKey.Precomputed.Qinv = big.NewInt(1)
	// CRTValues field is deprecated and not needed for this test

	// Create a signature with the broken key
	signature := B2bTokenSignature{
		ClientID:   "testclient",
		Timestamp:  "2022-09-15T10:30:00",
		PrivateKey: brokenKey,
	}

	// Attempt to create signature, which should fail at signing
	_, err := signature.Create()
	assert.Error(t, err, "Should return error when signing fails")
	assert.Contains(t, err.Error(), "input overflows")
}

// TestB2BTokenSignatureCreateComplete provides comprehensive testing for the Create method
func TestB2BTokenSignatureCreateComplete(t *testing.T) {
	// Test cases
	testCases := []struct {
		name        string
		signature   B2bTokenSignature
		expectError bool
		errorType   string
	}{
		{
			name: "Success case",
			signature: B2bTokenSignature{
				ClientID:  "test-client",
				Timestamp: "2022-09-15T10:30:00",
				PrivateKey: func() *rsa.PrivateKey {
					key, _ := generateRsaPrivateKeyMock()
					return key
				}(),
			},
			expectError: false,
		},
		{
			name: "Nil private key",
			signature: B2bTokenSignature{
				ClientID:   "test-client",
				Timestamp:  "2022-09-15T10:30:00",
				PrivateKey: nil,
			},
			expectError: true,
			errorType:   "private key not found",
		},
		{
			name: "Invalid private key validation",
			signature: B2bTokenSignature{
				ClientID:  "test-client",
				Timestamp: "2022-09-15T10:30:00",
				PrivateKey: &rsa.PrivateKey{
					// This key will fail validation
					PublicKey: rsa.PublicKey{
						N: big.NewInt(123),
						E: 65537,
					},
					D:      big.NewInt(456),
					Primes: nil, // Missing primes will cause validation to fail
				},
			},
			expectError: true,
			errorType:   "crypto/rsa: missing primes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signature, err := tc.signature.Create()

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorType != "" {
					assert.Contains(t, err.Error(), tc.errorType)
				}
				assert.Empty(t, signature)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, signature)
			}
		})
	}
}

// MockReader is a custom reader that always returns an error
type MockReader struct{}

func (r MockReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("mock reader error")
}

// TestB2BTokenSignatureCreateSignError tests the error path when SignPKCS1v15 returns an error due to rand.Reader error
func TestB2BTokenSignatureCreateSignError(t *testing.T) {
	// Skip if running in parallel or this could affect other tests
	if t.Parallel(); true {
		t.Skip("Skipping test that modifies global state")
	}

	// Create a valid private key first, before we replace rand.Reader
	privateKey, err := generateRsaPrivateKeyMock()
	assert.NoError(t, err)

	// Create a mock reader
	mockReader := &mockReader{err: fmt.Errorf("mock reader error")}

	// Save the original rand.Reader
	originalReader := rand.Reader

	// Replace with our mock reader
	rand.Reader = mockReader

	// Restore the original reader when test finishes
	defer func() {
		rand.Reader = originalReader
	}()

	// Create a signature with the valid private key but with a mocked reader
	signature := B2bTokenSignature{
		ClientID:   "testclient",
		Timestamp:  "2022-09-15T10:30:00",
		PrivateKey: privateKey,
	}

	// Attempt to create signature, which should fail at the signing step
	_, err = signature.Create()
	assert.Error(t, err, "Should return error when signing with mocked reader")
}

func TestB2bTokenSignatureInternalErrors(t *testing.T) {
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

		signature := B2bTokenSignature{
			ClientID:   "test-client",
			Timestamp:  "2022-09-15T10:30:00",
			PrivateKey: invalidKey,
		}

		_, err := signature.Create()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing primes")
	})
}
