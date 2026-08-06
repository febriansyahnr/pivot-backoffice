package encryption_test

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/encryption"

	"github.com/stretchr/testify/require"
)

func TestCryptoProviderDecryptHybrid(t *testing.T) {
	provider := NewCryptoProvider()

	payloadString := `{"encryptedKey":"Iw9UvLSxahaK7bePjXxLEdb5iaYnUTRtxWh7iYjiFLTDPibxqlbtETBkUTqgNJRNjxpRO3SbLmXFGgBIpj5n1vdyxhjBKExqo7thY8Ggk8vP2KiYNZhiLsNuZkeku8xa6YBX9hdZSwUjYfJEpJp61LKO9lUNwXlfFMO18NaoNnZwmseQYuWwB0ujKsfCUZ9KChvqlMUsec+Ro6EyOjB3u4iJGyPLE2CJ44BtPVZ6Fr8cW4AbWq7Ete9Cchy5A2M3tVChrtMCvaGsJ3dTcMWHtdpNA6kV8lhdE8plgLzdUQblCyYN1iUOMBGyq7nM3xK27m8FAS1Xqxgxo7aCr0+e3Q==","nonce":"PkFnXI6pjPDHoGJV","ciphertext":"FqYsUPRBqDQcgdK6yTV6Ef/vRzDg8IiAISysqw=="}`

	request := DataEncryption{}
	require.NoError(t, json.Unmarshal([]byte(payloadString), &request))

	privateKeyPem := `-----BEGIN PRIVATE KEY-----
MIIEugIBADANBgkqhkiG9w0BAQEFAASCBKQwggSgAgEAAoIBAQDBgrQOp+3VZ7xO
t6BTU0Jjr61T3jBk68MEfdgRAwE7rUGpvzWlAPRuXAIgY3IXNzb+yU9SshD+q1p7
Tg7j6bRs8yZne86Hh+FbXXUWPo5jSbq2NFhkz1EGCYsDXiSeegyuMDaHFbDRY1NQ
LZ5ia3TVJk+4Cf992IkOIUH4ePqHOmmeOA1A2tKiMzKYm/apuuYWCJbWxxhTWZNi
p60aPo0HiIZJAshbO14SOVeDNk4444v2/3VKARzvE183Yu9GDZQc85b4+wBNkArc
C7vxeZGHTbOimR78erWZpdG2NpRf7GZUhvIHxk02jqiSiUN2v9opFQqeE9TEJfGm
3dNhu6+JAgMBAAECgf9ZkqbGIfV1Uw9XYKhV6bRRcIBK0g6UqI0dByN8vGdVuF0s
tdWfC1IZw304gM/O70AFsCmHneU//RFlAjziQsvBosukGyr+kWc/Y4NPSKWgUACd
Vp2UDGL06HXwmSNaOCSmfKrA3Ml8Iv3tnzAxXow2HgXIqtgY6KKIZ9yhp15QOWuX
hnPmqe659lDbQAa3WPacUgicSOOC0o8Dp3G+UnbGs7UALphXyBNSrtfvHr4BpClV
qMmx6wFohIlVOwXzvAQVOzRC0m4MpyVwfPSJMRclnSIWvNPhjcVnMh+DffBVxKe2
tOko2yQCheuZM0DilMfKWjFfCf7l9MYnyKrrLmkCgYEA4al/mddJm7sErus7GWXp
X5mk5fiEzFBXCw+DAiqaq7w+eTzAzbrg+hflM91hpRtjLGpiw7J8Prk/D1EfSSyO
3ryxXXLg0K4jxI1Uc8exNLVze8aZ5uv00hCgg0Sn3jv/IbBXr62Uone8CxHFS8jE
Z/2iErDSXG2rY9BCJATtSF0CgYEA24apO2cnt64pBbB59i584mQDPoT/VAOiPznW
/CuN1B6k0kfoHjfX3c+kytIyQSVQe6+QjBnGUz5rp7c9en+N5mDOBsXGmcHPOjnn
CXYoOYXMpq3UMVO/46DbH4s7pCa4XHG3F284lFoDX6u2PslX+6DmlSYA8ohFQOfE
CJQ4oR0CgYAtpcTbENKh/uXGoGzXCWd44DKcFnZ+ge3pndypbobVIIIesixqMVhb
HsRNhoW/CVg5XtfVsGAzq/NWnNlQSwQniFH0jk1tyRwRIWmo9gchm2bd7eGp9acT
ayudAiFW8hn87Zf/QISljMTsFE8tslIQmxzS3RPggIq/6RvH/3skPQKBgEQ7fMpb
67ppxZJhIedk16g+UcvS5tGkN3/TaIEEwJaX178MXpdV4CCvc5ce8kPRZ0yqaxFA
yaYCFtAQYml60A41NJRiULJlzRVZ/few5BvM/KkqCnQyhcgorTMGwcjpyA/jwHbm
OP3TZI0OAB1P06sAfesJ3u2DcZMU9pd8CoX5AoGAKKd2HSuMw1+6cEBdUpbXirNf
boR0/b7/cTumL3thLJOBbEe9StvBdIhmrm/dgKKMOLyp/DbFw+PujRJwT4JjlMxf
7MtCFLgfLkhTqZGCVY7H/20TVtIRxEuwz95jyMoYJPxeQYIIGMXu/YFmmhu81fJX
PV7XrygiZSNhzZBK6pE=
-----END PRIVATE KEY-----
` // NOSONAR
	privateKeyPkcs8 := `MIIEugIBADANBgkqhkiG9w0BAQEFAASCBKQwggSgAgEAAoIBAQDBgrQOp+3VZ7xOt6BTU0Jjr61T3jBk68MEfdgRAwE7rUGpvzWlAPRuXAIgY3IXNzb+yU9SshD+q1p7Tg7j6bRs8yZne86Hh+FbXXUWPo5jSbq2NFhkz1EGCYsDXiSeegyuMDaHFbDRY1NQLZ5ia3TVJk+4Cf992IkOIUH4ePqHOmmeOA1A2tKiMzKYm/apuuYWCJbWxxhTWZNip60aPo0HiIZJAshbO14SOVeDNk4444v2/3VKARzvE183Yu9GDZQc85b4+wBNkArcC7vxeZGHTbOimR78erWZpdG2NpRf7GZUhvIHxk02jqiSiUN2v9opFQqeE9TEJfGm3dNhu6+JAgMBAAECgf9ZkqbGIfV1Uw9XYKhV6bRRcIBK0g6UqI0dByN8vGdVuF0stdWfC1IZw304gM/O70AFsCmHneU//RFlAjziQsvBosukGyr+kWc/Y4NPSKWgUACdVp2UDGL06HXwmSNaOCSmfKrA3Ml8Iv3tnzAxXow2HgXIqtgY6KKIZ9yhp15QOWuXhnPmqe659lDbQAa3WPacUgicSOOC0o8Dp3G+UnbGs7UALphXyBNSrtfvHr4BpClVqMmx6wFohIlVOwXzvAQVOzRC0m4MpyVwfPSJMRclnSIWvNPhjcVnMh+DffBVxKe2tOko2yQCheuZM0DilMfKWjFfCf7l9MYnyKrrLmkCgYEA4al/mddJm7sErus7GWXpX5mk5fiEzFBXCw+DAiqaq7w+eTzAzbrg+hflM91hpRtjLGpiw7J8Prk/D1EfSSyO3ryxXXLg0K4jxI1Uc8exNLVze8aZ5uv00hCgg0Sn3jv/IbBXr62Uone8CxHFS8jEZ/2iErDSXG2rY9BCJATtSF0CgYEA24apO2cnt64pBbB59i584mQDPoT/VAOiPznW/CuN1B6k0kfoHjfX3c+kytIyQSVQe6+QjBnGUz5rp7c9en+N5mDOBsXGmcHPOjnnCXYoOYXMpq3UMVO/46DbH4s7pCa4XHG3F284lFoDX6u2PslX+6DmlSYA8ohFQOfECJQ4oR0CgYAtpcTbENKh/uXGoGzXCWd44DKcFnZ+ge3pndypbobVIIIesixqMVhbHsRNhoW/CVg5XtfVsGAzq/NWnNlQSwQniFH0jk1tyRwRIWmo9gchm2bd7eGp9acTayudAiFW8hn87Zf/QISljMTsFE8tslIQmxzS3RPggIq/6RvH/3skPQKBgEQ7fMpb67ppxZJhIedk16g+UcvS5tGkN3/TaIEEwJaX178MXpdV4CCvc5ce8kPRZ0yqaxFAyaYCFtAQYml60A41NJRiULJlzRVZ/few5BvM/KkqCnQyhcgorTMGwcjpyA/jwHbmOP3TZI0OAB1P06sAfesJ3u2DcZMU9pd8CoX5AoGAKKd2HSuMw1+6cEBdUpbXirNfboR0/b7/cTumL3thLJOBbEe9StvBdIhmrm/dgKKMOLyp/DbFw+PujRJwT4JjlMxf7MtCFLgfLkhTqZGCVY7H/20TVtIRxEuwz95jyMoYJPxeQYIIGMXu/YFmmhu81fJXPV7XrygiZSNhzZBK6pE=` // NOSONAR
	request.PrivateKeyPEM = privateKeyPem

	plaintext, err := provider.DecryptHybrid(&request)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", string(plaintext))

	request.PrivateKeyPEM = ""
	request.PrivateKeyPKCS8 = privateKeyPkcs8
	plaintext, err = provider.DecryptHybrid(&request)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", string(plaintext))
}

func TestCryptoProviderDecryptAESCBC(t *testing.T) {
	provider := NewCryptoProvider()

	var (
		key16   = []byte("1234567890123456")                 // 16 bytes - AES-128
		key24   = []byte("123456789012345678901234")         // 24 bytes - AES-192
		key32   = []byte("12345678901234567890123456789012") // 32 bytes - AES-256
		validIV = []byte("abcdefghijklmnop")                 // 16 bytes
	)

	// encrypt applies PKCS7 padding and encrypts plaintext with AES-CBC.
	encrypt := func(key, iv, plaintext []byte) []byte {
		padding := aes.BlockSize - len(plaintext)%aes.BlockSize
		padded := make([]byte, len(plaintext)+padding)
		copy(padded, plaintext)
		for i := len(plaintext); i < len(padded); i++ {
			padded[i] = byte(padding)
		}
		block, err := aes.NewCipher(key)
		require.NoError(t, err)
		ciphertext := make([]byte, len(padded))
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
		return ciphertext
	}

	// encryptRaw encrypts data as-is with AES-CBC without applying PKCS7 padding.
	encryptRaw := func(key, iv, data []byte) []byte {
		block, err := aes.NewCipher(key)
		require.NoError(t, err)
		ciphertext := make([]byte, len(data))
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, data)
		return ciphertext
	}

	tests := []struct {
		name       string
		key        []byte
		iv         []byte
		ciphertext []byte
		wantResult []byte
		wantErrMsg string
	}{
		{
			name:       "success/AES-128 single block plaintext",
			key:        key16,
			iv:         validIV,
			ciphertext: encrypt(key16, validIV, []byte("Hello World!")),
			wantResult: []byte("Hello World!"),
		},
		{
			name:       "success/AES-192 key",
			key:        key24,
			iv:         validIV,
			ciphertext: encrypt(key24, validIV, []byte("Hello World!")),
			wantResult: []byte("Hello World!"),
		},
		{
			name:       "success/AES-256 key",
			key:        key32,
			iv:         validIV,
			ciphertext: encrypt(key32, validIV, []byte("Hello World!")),
			wantResult: []byte("Hello World!"),
		},
		{
			name:       "success/multi-block plaintext",
			key:        key16,
			iv:         validIV,
			ciphertext: encrypt(key16, validIV, []byte("This is a longer plaintext that spans multiple AES blocks!")),
			wantResult: []byte("This is a longer plaintext that spans multiple AES blocks!"),
		},
		{
			name:       "error/IV too short",
			key:        key16,
			iv:         []byte("tooshort"), // 8 bytes
			ciphertext: encrypt(key16, validIV, []byte("test")),
			wantErrMsg: "iv must be 16 bytes, got 8",
		},
		{
			name:       "error/IV too long",
			key:        key16,
			iv:         []byte("12345678901234567"), // 17 bytes
			ciphertext: encrypt(key16, validIV, []byte("test")),
			wantErrMsg: "iv must be 16 bytes, got 17",
		},
		{
			name:       "error/empty ciphertext",
			key:        key16,
			iv:         validIV,
			ciphertext: []byte{},
			wantErrMsg: "ciphertext length must be a non-zero multiple of aes block size",
		},
		{
			name:       "error/ciphertext length not multiple of block size",
			key:        key16,
			iv:         validIV,
			ciphertext: make([]byte, 15), // 15 is not a multiple of 16
			wantErrMsg: "ciphertext length must be a non-zero multiple of aes block size",
		},
		{
			name:       "error/invalid AES key size",
			key:        []byte("badkey"), // 6 bytes - not a valid AES key size
			iv:         validIV,
			ciphertext: make([]byte, 16),
			wantErrMsg: "new aes cipher",
		},
		{
			name: "error/PKCS7 padding value is zero",
			key:  key16,
			iv:   validIV,
			// all-zero plaintext: last byte = 0x00, padLen = 0 → invalid
			ciphertext: encryptRaw(key16, validIV, make([]byte, 16)),
			wantErrMsg: "remove pkcs7 padding",
		},
		{
			name: "error/PKCS7 padding value exceeds block size",
			key:  key16,
			iv:   validIV,
			// last byte = 17 (> aes.BlockSize = 16) → invalid
			ciphertext: encryptRaw(key16, validIV, append(make([]byte, 15), byte(aes.BlockSize+1))),
			wantErrMsg: "remove pkcs7 padding",
		},
		{
			name: "error/inconsistent PKCS7 padding bytes",
			key:  key16,
			iv:   validIV,
			// last byte = 4, but data[12..14] = {0x01, 0x02, 0x03} ≠ 0x04 → invalid
			ciphertext: encryptRaw(key16, validIV, []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04,
			}),
			wantErrMsg: "remove pkcs7 padding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.DecryptAESCBC(tt.key, tt.iv, tt.ciphertext)

			if tt.wantErrMsg != "" {
				require.ErrorContains(t, err, tt.wantErrMsg)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestCryptoProviderEncryptPKCS7(t *testing.T) {
	provider := NewCryptoProvider()

	validCertPEM := `-----BEGIN CERTIFICATE-----
MIIDazCCAlOgAwIBAgIUdWGdcs98rVBBkKwAGf9xHmzvYMgwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yNjAyMjAwNzE3MTVaFw0yNzAy
MjAwNzE3MTVaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQCtLfNoWW7JJo2bwep1o7CHv0D69H/b+hasYdVfma+0
mvWIV8/iELNXYN61Ol1h4oXdNfrXJgzsEvjKqvBlaxysTamBwFPEMDLdoq+76IkK
cgMAeXNFn9cFdntYcM0o2b7u+CbIuBkcQOKm9AKGUiWx9iK5a2tORk6zLnEjIN+b
lNa4NaBOo1bxY6xb5Bi9iZDvzHgiZGQC6JdQzf1xuA9VBS7fSMyTR2fVmfKpYc7v
NlnwS5psCNyY7wkUG/TsMsLfPiP+ADcAg5gXLZU/0OXG5NayIHoQtstI8cmtv6+l
O3ixub0vBISVib0V7GtftiEDwuiBN2LDiZhgoWtcrHNPAgMBAAGjUzBRMB0GA1Ud
DgQWBBTn69NEHLXCok4eIs5jxJH5Co5J0TAfBgNVHSMEGDAWgBTn69NEHLXCok4e
Is5jxJH5Co5J0TAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCF
RaEren46WJI1klKQzjXHbpO75/9sB9C/e9jZeF0GnVCBIIGYITOsb+G/EHNxmCOH
qEUsb3u/4WYi+/3RQ+I16sDnVYiUQF7bG0C8K8s2vk5Lv40xzDuNDFGTL79Xi7DZ
0EDDho7cnsXYTExFUZXWQa8nzCCKGA0xXDwWkLP2CHDolJNI1AmUS1sHl1QZI7e7
8+zSSX+zsrTRmAgLYhcupGnYt5JK5yMszm7VqU7FU67niBK/64W3woAw8zD8AGsK
gKZAGmjOK+4V23VfjsVs5dlTz+q1HytzV2jWEuAJ+uWPki6UpdMETC89ZApqbC5e
5pnHCRiBKU0Nkco1mLn4
-----END CERTIFICATE-----
` // NOSONAR

	tests := []struct {
		name            string
		certPem         []byte
		plaintext       []byte
		wantErrContains string
	}{
		{
			name:      "success/valid cert and plaintext",
			certPem:   []byte(validCertPEM),
			plaintext: []byte("Hello World!"),
		},
		{
			name:      "success/multi-block plaintext",
			certPem:   []byte(validCertPEM),
			plaintext: []byte("This is a longer plaintext that spans multiple blocks and contains more content!"),
		},
		{
			name:            "error/empty certPem",
			certPem:         []byte{},
			plaintext:       []byte("Hello World!"),
			wantErrContains: "failed to decode PEM block",
		},
		{
			name:            "error/invalid PEM format",
			certPem:         []byte("not a pem block"),
			plaintext:       []byte("Hello World!"),
			wantErrContains: "failed to decode PEM block",
		},
		{
			name: "error/PEM with invalid certificate bytes",
			certPem: pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: []byte("invalid der bytes"),
			}),
			plaintext:       []byte("Hello World!"),
			wantErrContains: "parse PKIX public key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.EncryptPKCS7(tt.certPem, tt.plaintext)

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
				require.Empty(t, result)
			} else {
				require.NoError(t, err)
				decoded, err := base64.StdEncoding.DecodeString(result)
				require.NoError(t, err, "result must be valid base64")
				require.NotEmpty(t, decoded)
			}
		})
	}
}
