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

func TestCreateSender(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore CreateSender service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{UUID: uuid.New()}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.CreateSender(context.Background(), &xbModel.CreateSenderRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetListSender(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore GetListSender service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetListSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetListSenderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetListSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetListSenderRequest"),
				).Return(&xbCoreProcessorModel.PaginationData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.GetListSender(context.Background(), &xbModel.GetListSenderRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSenderById(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore GetSenderById service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetSenderById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetSenderByIdRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetSenderById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetSenderByIdRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.GetSenderById(context.Background(), &xbModel.GetSenderByIdRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateSender(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore UpdateSender service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("UpdateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.UpdateSenderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("UpdateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.UpdateSenderRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.UpdateSender(context.Background(), &xbModel.UpdateSenderRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeactivateSender(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore DeactivateSender service error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("DeactivateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetSenderByIdRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("DeactivateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetSenderByIdRequest"),
				).Return(&xbCoreProcessorModel.CreateSenderData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo)
			_, err := svc.DeactivateSender(context.Background(), &xbModel.GetSenderByIdRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
