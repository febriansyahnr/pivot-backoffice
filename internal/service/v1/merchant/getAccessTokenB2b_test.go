package merchant

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	encryptMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantService_GetAccessTokenB2b(t *testing.T) {
	clientID := "client-id"
	clientSecret := "client-secret"

	vaultTransit := vaultMock.NewIVaultTransit(t)

	testCases := []struct {
		Name      string
		IsSuccess bool
		MockSetup func(
			mockRepo *mockMerchant.IMerchantRepository,

			jwtMock *mockJWT.IJwt,
		)
	}{
		{
			Name:      "SUCCESS: Get access token b2b",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository, jwtMock *mockJWT.IJwt) {
				mockRepo.On(
					"GetMerchantAuthByMerchantId", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.MerchantAuth{
					Secret:     clientSecret,
					MerchantID: clientID,
				}, nil)

				mockRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchant.Merchant{UUID: clientID}, nil)

				jwtMock.On(
					"GenerateMerchantToken",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return("valid-access-token", nil)
			},
		},
		{
			Name:      "ERROR: Get merchant auth by merchant id",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository, _ *mockJWT.IJwt) {
				mockRepo.On(
					"GetMerchantAuthByMerchantId", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, assert.AnError)
			},
		},
		{
			Name:      "ERROR: Merchant auth not found",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository, _ *mockJWT.IJwt) {
				mockRepo.On(
					"GetMerchantAuthByMerchantId", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, nil)
			},
		},
		{
			Name:      "ERROR: Decrypt merchant auth",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository, _ *mockJWT.IJwt) {
				mockRepo.On(
					"GetMerchantAuthByMerchantId", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.MerchantAuth{Secret: "vault:v1:...", SecretVersion: 1}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
		},
		{
			Name:      "ERROR: Merchant error repo",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository, jwtMock *mockJWT.IJwt) {
				mockRepo.On(
					"GetMerchantAuthByMerchantId", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.MerchantAuth{
					Secret:        "vault:v1:...",
					SecretVersion: 1,
					MerchantID:    clientID,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Return(&vault.DecryptResponse{Plaintext: []byte(clientSecret)}, nil)
				mockRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			Name:      "ERROR: Merchant not found",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository, jwtMock *mockJWT.IJwt) {
				mockRepo.On(
					"GetMerchantAuthByMerchantId", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.MerchantAuth{
					Secret:     clientSecret,
					MerchantID: clientID,
				}, nil)

				mockRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(nil, nil)
			},
		},
		{
			Name:      "ERROR: Failed to generate merchant token",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository, jwtMock *mockJWT.IJwt) {
				mockRepo.On(
					"GetMerchantAuthByMerchantId", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.MerchantAuth{
					Secret:     clientSecret,
					MerchantID: clientID,
				}, nil)

				mockRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchant.Merchant{UUID: clientID}, nil)

				jwtMock.On(
					"GenerateMerchantToken",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return("", errors.New("invalid merchant token"))

			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchantRepo := mockMerchant.NewIMerchantRepository(t)
			userRepo := mockUser.NewIUserRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			jwtMock := mockJWT.NewIJwt(t)
			accountSvc := mocks.NewIAccountService(t)

			tc.MockSetup(merchantRepo, jwtMock)
			svc := New(merchantRepo, loggerMock, userRepo, jwtMock, nil, nil, WithAccountService(accountSvc), WithVaultTransit(vaultTransit))

			response, err := svc.GetAccessTokenB2b(context.Background(), clientID, clientSecret)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}

func TestGetSNAPAccessTokenB2B(t *testing.T) {
	jwt := mockJWT.NewIJwt(t)
	logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	repo := mockMerchant.NewIMerchantRepository(t)
	encrypt := encryptMock.NewICrypto(t)

	service := New(repo, logger, nil, jwt, nil, encrypt)

	const (
		merchantId  = "3fc96de8-f65e-4b16-90a1-e2a00d1bae29"
		publicKey   = "ESzHH+A2uIDx8wtfodE2aCJaW1yA2e5tn2cwpBMFXJgBrg+47norHXB7sPzHzIOlR6kgQxy8qkxT89WBZ2WVNZRSlXpPPzwg4uP6tdZBGducKfaCNpMsCpR3QPJLsQzOOahhqabJh28dH5+oMq1f8QP3mott5fBrQoiXXJX1KTtRLd0hHJXWPNIS7oWB6caNPCAwpFJpHUdsP2ghsFdTkCNAIncexf5gWlrNPIZr1tQjAb6GGnWyLAPdTr55BdLPaQ63Ir5K3JAu9PJKa/3TBM4cub+xcFi+2KNMJhhR/xwfuzEtt0E+OeID9+w9dY85PhUiB4a+3aPGnqo5XrJ8TdAOz8bmuj37699670SSJsO8jpfc7YD5EUQVrnwnKm9HfP0gSC8AU1qoQBZDnqBIYC+5Hs0nCVkL2FOMx3V7dNKXZkhS/ARB7UQQYiyAAG5zGDCAKJcMfoYjpwiGAoiyVl5iVi7qbnpASDD88eLJVO1ND6YvAqJyBQanNWzYA1ad5P8WcpnRFkSiyepLsdkb1ltulTngpXgX/gUBbYXKGuozLMZiVgs0QVoWMrUz24WTuCmSbMFs2TMofAfjvnocTaFPZcfM23XPlTxVENy5rSOPJe9iGv0dnNfR/Stxwg=="
		privateKey  = "s60wPu2nztfqRMEMcaI44ZhoKofrORuz4a3iD6VmLd92hm3gf7R5KUB4kq0HYWA4iQ9cWpvJZdDp2r/irYr/WPM7a1+p9JdF00HRg2Aw/BvGzXJbProNiCsrYQP8zbzj6NKJpxxPwcgzJJzqb6dWwn1RaiIubl3mWiVlsXTjiErlSv8YfUsl24fDwsRMedPZ5TGnUk1qgWU81wW32M/ZkhC8psIjSlCjMe0XQKBuV0+cCwmBrhp2tt5e7uqfHYE7sqJJXzZ/va/tMQY7F9CXdx/wAOhQ95GEtly/7KD4LMAsHW7khLyGnYq/EI/RVM/toScnDKCsaqF3OsBpYBQt1rzmHfPae5VtNpVv0YncMNrpK9EOYDgDwsx0pX68xDkFAxhBp6iOsnBCt3tJ/BRR7yu5erkjvkiYqsauR4q0HG6s/ayWIawVnzMG1BXZt8HqPOdPVCRudWVkp523cDt8OvNn1UvCWW492ox2lm2QHFyFNUC8upP6PIpa/XOvt45cbfFCsCuFHLcmLpOJ9TpVwbnWegLIH/7O4t1hBD3Y3TYV2BI1/y+s4xKwO8x6DPadkDPswpM5DxhlHWvoKSUDrTLU3YvVkeh2qPSVNRICGV8pMFE06Ia+pJxB95NqVf1lpfBxTZN+YJL+oouro2myk0E33CD7EJ3zwzDA+rhV5GRkK+HU5rAd105xDwulk5vLgsjYKRQFqwOdmX2C7S7hSzgIgWSfXER7cRoWy7GA4i4nsaFp3sjLYXVsAdf9i9judGyz1Yhv5U4U6r1wpjXrhLKBKLI104JYxzKpxkmtWWrli44w7Hv1aOwJVwBEPgmG5EM9MdhM/zsUbtuh3V5DPfMKM921qGuhvfD0IKiVkDBKMmMmo0W+zdavTel3uuL/FMquEw5JGFw3GEuO3YxvY2zTT0es7YqY1zyS/3oUN2yUkZIskMKR1QSwkWTRni3GwyT7lj++3ErHRxS5v+z5BYvQ94schyUraXMcbmvWnnVO9iRH1BS8csL/hIux3AtJtfiWh8EIc/Vf3ZqGwABvyOVB5AIl+eMAtkjKq0MW0eOEwtCjJlzl03g4vNSgIP4shtbZN8LCkO+RSkUddZ3F74HUsuyVXjEOhHmW+qSXGQA1nI2UoFWNctUfjk9w1TCby70atNH/fdrtWCEFx8Kl0PvDZZoRtBAZELOdkRWL1EUwob1vapFFrFMXV51vvB+zMkYWcE1vsNkOd1ZGoSt0R4G1DLeT5dzYZCI8qxrgs7adrVAL6zeV6dnY/tPGTbVpSQqMURXKwmJxZWF7el/HUIxQ6Zx+Qa1V4OQm3GY3m8Oy9Ib/YBOlCNppsivdEhiGoTZPpvvxt+6wOCiFgaTM6aSTIwl9/v2y98rMd/5LzKE277naLppx4sQk9aH0+XzHU/OIW2ipE/orhGgMFx9igi0Ecym/Fh1PtfjmQKOZWeHTvlStmTfUBE3YO3g8boqhAD5aKVITAA2oOz/vFIstGp+GT/Hv39ziROuBPJ95bmX5bKDgpoqX4G3vjr/j36jHPX8RQpia/OiLceA9mvlyTn73HzV7XF/9BPXpqk6zTS6jbrb5BGxxlN0QOjo97h4j6wIhCwnMMryq9Z1ralIAGTi/so0UAmYyBANK7Wxkpf6vw1QQ84WbV42P40EJnYKqu2uoevnuDbo27DiF5c/kwJ5G8nilg+Kbua8RMDq2UHoFd6LVioZ664KY88RB1yAa8Yrzv5JmfyFQo3bxhiBThCHRJC+OdbwY2VB9mHdkCeOfy+tcPIfaL9fwD2fjnuHFFyZGCVjtGNeHd+nmoSxIyLSgfXDS2rbdixoPp0vkij79HRySPujE7/gBJAp/UcCxRSXWq0d4fKFWnefg3IrRsGv+QmL8FNrfhuzazn1T0Mj/ctf2ZbYFZbvLKPVf1hvMiNgI+lOGyV3RhEsH1Avr/jS/DUYgAEpMYrr2CHv7YdUUG806ChxT1E5QYbBJunfp4EMXWMNZ08uSjFyR7e8vK7GjJOsfgRppqtbzkGSfgejXdK9kEfNbrnpYeNEHD/7YEWPLR4XYf+r0lHeKEEKfExK3Ek3rBua5JBK4k+Vhp1wkzekdfTMwZRfMw/KPgGKL1bYuueUQlWf8zTpZxZJRhnMTfU8RSKw/DI8hURv8u+qoy7F1YsGOER37+gdLeXreHbLKeU91GE1eSLJ0AOM8vpKrSbzbIY+UWZ2CWvIZxhWpZzmmMeSpRsS0Z+ZWDXnj8HwR7HcK2O+SEYu9r+OJVgfDWs7koVRs143eeMry32DyREQBHIoVWIEWQFBY5EN7Tr4Y+A=="
		accessToken = "example-access-token"
	)
	e := encryption.New()

	merchantUUID, _ := uuid.Parse(merchantId)
	secretKey := e.SecretKeyFromUUID(merchantUUID)

	encrypt.On("SecretKeyFromUUID", merchantUUID).Return(secretKey)

	rawPublicKey, err := e.Decrypt(publicKey, secretKey)
	require.NoError(t, err)

	rsaPublicKey, err := e.ToPublicKey(*rawPublicKey)
	require.NoError(t, err)

	tests := []struct {
		name       string
		request    *merchant.SNAPAccessTokenB2BReq
		setupMock  func()
		wantErr    error
		wantResult *merchant.SNAPAccessTokenB2BResp
	}{
		{
			name: "ERROR:Get merchant auth data",
			setupMock: func() {
				repo.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrUnauthorized, fmt.Errorf(constant.UnauthorizedSnapFmt, constant.HeaderXClientKey)),
		},
		{
			name: "ERROR:Empty private and public key",
			setupMock: func() {
				repo.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&merchant.MerchantAuth{}, nil)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrUnauthorized, fmt.Errorf(constant.UnauthorizedSnapFmt, constant.HeaderXClientKey)),
		},
		{
			name: "ERROR:Decrypt merchant public key",
			setupMock: func() {
				repo.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.MerchantAuth{
					MerchantID:        merchantId,
					MerchantPublicKey: sql.NullString{String: publicKey},
					SnapPrivateKey:    sql.NullString{String: privateKey},
				}, nil)

				encrypt.On(
					"Decrypt", constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Decrypt merchant public key",
			setupMock: func() {
				repo.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.MerchantAuth{
					MerchantID:        merchantId,
					MerchantPublicKey: sql.NullString{String: publicKey},
					SnapPrivateKey:    sql.NullString{String: privateKey},
				}, nil)

				encrypt.On("Decrypt", publicKey, secretKey).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Convert to public key",
			setupMock: func() {
				encrypt.On("Decrypt", publicKey, secretKey).Return(rawPublicKey, nil)

				encrypt.On("ToPublicKey", *rawPublicKey).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Invalid public key",
			setupMock: func() {
				encrypt.On("ToPublicKey", *rawPublicKey).Once().Return(&rsa.PublicKey{N: &big.Int{}, E: 0}, nil)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrUnauthorized, fmt.Errorf(constant.UnauthorizedSnapFmt, constant.HeaderXSignature)),
		},
		{
			name: "ERROR:Merchant repo error",
			setupMock: func() {
				encrypt.On("ToPublicKey", *rawPublicKey).Return(rsaPublicKey, nil)

				repo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrFindMerchant),
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrUnauthorized, constant.ErrMerchantNotFound),
		},
		{
			name: "ERROR:Generate merchant token",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID: merchantId,
				}, nil)

				jwt.On(
					"GenerateMerchantToken", constant.ValueCtxMockType(), constant.StringMockType(), merchantId,
				).Once().Return("", constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				jwt.On(
					"GenerateMerchantToken", constant.ValueCtxMockType(), constant.StringMockType(), merchantId,
				).Return(accessToken, nil)
			},
			wantResult: &merchant.SNAPAccessTokenB2BResp{
				ResponseCode:    "2007300",
				ResponseMessage: "Successful",
				AccessToken:     accessToken,
				TokenType:       "Bearer",
				ExpiresIn:       "900",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.request == nil {
				test.request = &merchant.SNAPAccessTokenB2BReq{
					ClientId:  merchantId,
					Timestamp: "2024-08-07T11:21:00+07:00",
					Signature: "BTdxROPJrA4eJiCW2dm+eEKY90Yep6kT23Ggx8TfxYOdS4P1qT4Br+euZmmBzCI1t+ybXPogPIgD7a9F6MP+DhQ1demLMxWZCyCjWkvglPx8bPRJn/FtQKbuEynHpWZg3fbfWIaARoMJc95m17rAOBCsLV922ycXnEsjtBcb+YPBZwMxEqoroDt7nSu2IORD+lzHh7hgnT6jls5QKFkPCn6Vyo/4VHAAbw/0zwbqakXyqAm5Nht+wF2fSyKKbJ6q/5/nQGHMVp+tbhV0nqty8aGNltFmt3R+M0tBdovPuQCL41nbguDJ0dQZBm1J89ykNNKmfBpP80tXtwJBPpC8/g==",
				}
			}
			test.setupMock()

			result, err := service.GetSNAPAccessTokenB2B(context.Background(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
