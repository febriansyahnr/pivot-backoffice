package userLoggedInDeviceService_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userLoggedInDeviceModel "github.com/paper-indonesia/pivot-backoffice/internal/model/userLoggedInDevice"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/userLoggedInDevice"
)

func TestValidate(t *testing.T) {
	repo := repositoryMocks.NewIUserLoggedInDeviceRepository(t)
	logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	cfg := &config.Config{UserOTPConfig: config.UserOTPConfig{UserLoginRememberInMinute: 10}}
	svc := New(cfg, nil, logger, Repositories{UserLoggedInDeviceRepo: repo})

	deviceId := "deviceId"

	testCases := []struct {
		name             string
		mocksSetup       func()
		wantErr          bool
		deviceIdentifier string
		isRemember       bool
	}{
		{
			name: "Error: Empty Device identifier",
			mocksSetup: func() {
				// empty setup
			},
			wantErr: true,
		},
		{
			name: "Error: FindByUserAndDevice error repo",
			mocksSetup: func() {
				repo.On("FindByUserAndDevice", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr:          true,
			deviceIdentifier: deviceId,
			isRemember:       true,
		},
		{
			name: "SUCCESS: Valid",
			mocksSetup: func() {
				validMetadata := `{"isRemember":true, "rememberUntil":"` + time.Now().UTC().Add(time.Duration(24)*time.Hour).Format(constant.DatetimeRFC3339Format) + `"}`
				repo.On("FindByUserAndDevice", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(&userLoggedInDeviceModel.UserLoggedInDevice{
					AdditionalInfo: &validMetadata,
				}, nil)
			},
			wantErr:          false,
			deviceIdentifier: deviceId,
			isRemember:       true,
		},
		{
			name: "ERROR: Create",
			mocksSetup: func() {
				repo.On("FindByUserAndDevice", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, nil)

				repo.On("Create", constant.ValueCtxMockType(), constant.PtrUserLoggedInDeviceMockType()).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr:          true,
			deviceIdentifier: deviceId,
			isRemember:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mocksSetup()

			err := svc.Validate(context.Background(), uuid.NewString(), tc.deviceIdentifier, tc.isRemember)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
