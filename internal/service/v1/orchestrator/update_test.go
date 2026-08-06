package orchestrator_service

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateStatusAccountTransaction(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateStatusAccountTransaction",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			err := accTrxSvc.UpdateStatusAccountTransaction(ctx, "id", constant.StatusSuccess, nil, nil)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}

func TestUpdateTransactionTimestamp(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateTransactionTimestamp",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			err := accTrxSvc.UpdateTransactionTimestamp(ctx, "id", time.Now())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}

func TestUpdateStatusAccountTransactionByReferenceID(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateStatusAccountTransactionByReferenceID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "FAIL: Repository Error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateStatusAccountTransactionByReferenceID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			err := accTrxSvc.UpdateStatusAccountTransactionByReferenceID(ctx, "reference-id", constant.StatusSuccess, nil, nil)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}

func TestUpdateAdditionalInfoByID(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateAdditionalInfoByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("types.NullJSONText"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "FAIL: Repository Error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateAdditionalInfoByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("types.NullJSONText"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			additionalInfo := []byte(`{"key": "value"}`)
			err := accTrxSvc.UpdateAdditionalInfoByID(ctx, "id", additionalInfo)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}

func TestUpdateStatusAccountAndAdditionalInfoTransaction(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateTransactionsStatusAndAdditionalInfoByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("types.NullJSONText"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "FAIL: Repository Error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateTransactionsStatusAndAdditionalInfoByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("types.NullJSONText"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			additionalInfo := []byte(`{"key": "value"}`)
			err := accTrxSvc.UpdateStatusAccountAndAdditionalInfoTransaction(ctx, "id", constant.StatusSuccess, "reason_type", "reason_description", additionalInfo)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}

func TestUpdateProcessorAndReconReferenceByID(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateProcessorAndReconReference",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "FAIL: Repository Error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateProcessorAndReconReference",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			err := accTrxSvc.UpdateProcessorAndReconReferenceByID(ctx, "id", "processor_name", "processor_id", "recon_reference")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}

func TestUpdateTransaction(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateTransactionDetail",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "FAIL: Repository Error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository) {
				accTrxRepo.On(
					"UpdateTransactionDetail",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, nil)
			ctx := context.Background()
			err := accTrxSvc.UpdateTransaction(ctx, &orchestrator_model.UpdateTransactionRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}
