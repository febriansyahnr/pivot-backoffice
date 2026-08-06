package xbPayoutService_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
)

func TestCreateBeneficiary(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	beneficiaryAccountRepo := repositoryMock.NewIBeneficiaryAccountRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore CreateBeneficiary service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("CreateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateBeneficiaryRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Upsert on Beneficiary Account Repo error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("CreateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateBeneficiaryRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{UUID: uuid.New()}, nil)

				beneficiaryAccountRepo.On("Upsert",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				beneficiaryAccountRepo.On("Upsert",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*beneficiaryAccountModel.BeneficiaryAccount"),
				).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, beneficiaryAccountRepo, xbCoreProcessorRepo)
			_, err := svc.CreateBeneficiary(context.Background(), &xbModel.CreateBeneficiaryRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetListBeneficiary(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore GetListBeneficiary service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetListBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetListBeneficiaryRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetListBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetListBeneficiaryRequest"),
				).Return(&xbCoreProcessorModel.PaginationData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.GetListBeneficiary(context.Background(), &xbModel.GetListBeneficiaryRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBeneficiaryById(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore GetBeneficiaryById service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetBeneficiaryById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetBeneficiaryByIdRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetBeneficiaryById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetBeneficiaryByIdRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.GetBeneficiaryById(context.Background(), &xbModel.GetBeneficiaryByIdRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateBeneficiary(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore UpdateBeneficiary service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("UpdateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.UpdateBeneficiaryRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("UpdateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.UpdateBeneficiaryRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.UpdateBeneficiary(context.Background(), &xbModel.UpdateBeneficiaryRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeactivateBeneficiary(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore DeactivateBeneficiary service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("DeactivateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetBeneficiaryByIdRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("DeactivateBeneficiary",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetBeneficiaryByIdRequest"),
				).Return(&xbCoreProcessorModel.CreateBeneficiaryData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.DeactivateBeneficiary(context.Background(), &xbModel.GetBeneficiaryByIdRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
