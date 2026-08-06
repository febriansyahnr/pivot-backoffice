package encryption

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSecretFromUUID(t *testing.T) {

	var (
		uuidReq        uuid.UUID
		expectedResult string
	)

	testCases := []struct {
		desc      string
		mockSetup func()
	}{
		{
			desc: "valid uuid",
			mockSetup: func() {
				uuidReq = uuid.New()
				expectedResult = strings.ReplaceAll(uuidReq.String(), "-", "")
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			crypt := New()

			tc.mockSetup()

			result := crypt.SecretKeyFromUUID(uuidReq)

			assert.Equal(t, expectedResult, result)
		})
	}
}

func TestEncrypt(t *testing.T) {
	cryptFunc := Crypto{}

	testCases := []struct {
		desc      string
		plainText func() string
		key       func() string
		wantErr   bool
	}{
		{
			desc: "success encrypting key",
			plainText: func() string {
				privKey, err := cryptFunc.GenerateRandomPKCS8Key()
				if err != nil {
					t.Fatal(err)
				}

				return string(privKey)
			},
			key: func() string {
				merchantId := "e41d5858-bf37-48f0-a6fb-9e5e0bbb73bd"

				merchantUUID, err := uuid.Parse(merchantId)
				if err != nil {
					t.Fatal(err)
				}

				return cryptFunc.SecretKeyFromUUID(merchantUUID)
			},
			wantErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			crypt := New()

			res, err := crypt.Encrypt(tc.plainText(), tc.key())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				t.Logf("encryptCreated: %v", res)
				assert.NotEmpty(t, res)
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	cryptFunc := Crypto{}

	testCases := []struct {
		desc       string
		cipherText string
		secretKey  func() string
		wantErr    bool
	}{
		{
			desc:       "success decrypting key",
			cipherText: "F4mVAqX8uS2q8CL7bRaeIKjGTCpp6bUyBN2pcaihy3UqyG6cStuOpOqmiQOQ9zXXgwqOXWzQuybiViMTXuQGcYu9JMoBnpXmoATKbKn6fmCgTPCNgzHoGOkOcQQEH1b1I7BOmpaPUB6dk0wAM6ZbQC8yF+VA/Id3MalsAbhy9qQtN8A7uosrQBSxHBjWIBSqQPZ/llbiHBfuXZwzyTOt5HbJD3qzJ3yAJJ6cyac3F7qTOpskUeQy8gt2X1P7IT4bOfcS0XPqzR41OWPFUN/UI00sWrsY9C3vNrXCj29Jo/VnD5He88PtaG6RuR17rZ3PqoUqPEQ8Jkxs4rl8M5Jdc0qzP4njLZt9cTEvbMaA3niAj5qwHNgIeQtBY/fJ16HSMPpV94/N+f+0E3oQ33JEmsyppJ4QsWI/m6XZrvP8gKCeRwW5zBnhczE+su5daGzmcnPoOf5N5pgpFxo1U/iTqu9+3/4nXmoSmR7pM2mPRANNEtyaApgs+KHM3DFHrd4gqhBM1zb7W30xTAX9sx71vNKHF6tcV9hLhEk3n9gT9TfnI6vS1sl2TJ+t5pSwc1SSfYlAeKtchbvWq+cO4bY=",
			secretKey: func() string {
				merchantId := "e41d5858-bf37-48f0-a6fb-9e5e0bbb73bd"

				merchantUUID, err := uuid.Parse(merchantId)
				if err != nil {
					t.Fatal(err)
				}

				return cryptFunc.SecretKeyFromUUID(merchantUUID)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			crypt := New()

			res, err := crypt.Decrypt(tc.cipherText, tc.secretKey())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, res)
				assert.NoError(t, err)
			}
		})
	}
}

func TestEncryptPassword(t *testing.T) {
	tests := []struct {
		pass string
		want string
	}{
		{
			pass: "HelloJoko123!@#",
			want: "49ad825f7958dbc2b23e1466279620d8ef239f2e4f50863c43dc741bea96e816",
		},
		{
			pass: "Qwerty123!@#",
			want: "63a7c40efdc57f082e8ac9bcc1ef5b6f9e4d662da3951a9c92b8ad49e1c7dfe4",
		},
		{
			pass: "abc123",
			want: "6ca13d52ca70c883e0f0bb101e425a89e8624de51db2d2392593af6a84118090",
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, EncryptPassword(test.pass))
	}
}
