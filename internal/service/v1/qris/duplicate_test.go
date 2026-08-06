package qris_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	qrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	qrisService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/qris"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDuplicateRegistration(t *testing.T) {
	ctx := context.Background()
	mockLog, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	t.Run("Success", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Create test data
		sourceMerchant := &merchant.Merchant{
			UUID:       "source-merchant-id",
			ExternalId: "source-external-id",
			Name:       "Source Merchant",
		}

		targetMerchant := &merchant.Merchant{
			UUID:       "target-merchant-id",
			ExternalId: "target-external-id",
			Name:       "Target Merchant",
		}

		callbackDetailRaw := types.NullJSONText{}
		_ = callbackDetailRaw.Scan([]byte(`{"applymentCode": "123", "resultCode": "SUCCESS"}`))

		acquirerMerchantId := "acquirer-merchant-id"

		sourceRegistration := &qrisModel.Registration{
			Id:                       "existing-reg-id",
			ExternalId:               "source-external-id",
			Acquirer:                 "BNC",
			MerchantType:             "Merchant",
			AcquirerParentMerchantId: "parent-merchant-id",
			MerchantName:             "Source Merchant",
			MerchantShortName:        "SM",
			Status:                   constant.QrRegistrationStatusSuccess,
			AcquirerMerchantId:       &acquirerMerchantId,
			CallbackDetailRaw:        callbackDetailRaw,
			CallbackDatetime:         sql.NullTime{Time: time.Now(), Valid: true},
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(sourceMerchant, nil)
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "target-merchant-id").Return(targetMerchant, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "target-external-id").Return(nil, nil) // Target has no registration
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "source-external-id").Return(sourceRegistration, nil)
		mockQrisRepo.On("InitRegistration", mock.AnythingOfType("*context.valueCtx"), mock.MatchedBy(func(r interface{}) bool {
			return true // Match any registration object
		})).Return(nil).Run(func(args mock.Arguments) {
			reg := args.Get(1).(*qrisModel.Registration)
			assert.Equal(t, "target-external-id", reg.ExternalId)
			assert.Equal(t, "Target Merchant", reg.MerchantName)
			assert.Equal(t, constant.QrRegistrationStatusSuccess, reg.Status)
			assert.Equal(t, "acquirer-merchant-id", *reg.AcquirerMerchantId)
			assert.Equal(t, "system-duplicate", reg.CreatedBy)
		})

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		mockMerchantRepo.AssertExpectations(t)
		mockQrisRepo.AssertExpectations(t)
	})

	t.Run("Target Merchant Already Has Registration", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Create test data
		sourceMerchant := &merchant.Merchant{
			UUID:       "source-merchant-id",
			ExternalId: "source-external-id",
			Name:       "Source Merchant",
		}

		targetMerchant := &merchant.Merchant{
			UUID:       "target-merchant-id",
			ExternalId: "target-external-id",
			Name:       "Target Merchant",
		}

		existingTargetRegistration := &qrisModel.Registration{
			Id:                "existing-target-reg-id",
			ExternalId:        "target-external-id",
			MerchantName:      "Target Merchant",
			MerchantShortName: "TM",
			Status:            constant.QrRegistrationStatusSuccess,
		}

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(sourceMerchant, nil)
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "target-merchant-id").Return(targetMerchant, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "target-external-id").Return(existingTargetRegistration, nil)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), "target merchant already has a QRIS registration")
		mockMerchantRepo.AssertExpectations(t)
		mockQrisRepo.AssertExpectations(t)
	})

	t.Run("Source Merchant Not Found", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(nil, nil)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), constant.ErrMerchantNotFound.Error())
		mockMerchantRepo.AssertExpectations(t)
	})

	t.Run("Database Error When Finding Source Merchant", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Set expectations
		dbErr := errors.New("database error")
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(nil, dbErr)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), dbErr.Error())
		mockMerchantRepo.AssertExpectations(t)
	})

	t.Run("Target Merchant Not Found", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Create test data
		sourceMerchant := &merchant.Merchant{
			UUID:       "source-merchant-id",
			ExternalId: "source-external-id",
			Name:       "Source Merchant",
		}

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(sourceMerchant, nil)
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "target-merchant-id").Return(nil, nil)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), constant.ErrMerchantNotFound.Error())
		mockMerchantRepo.AssertExpectations(t)
	})

	t.Run("Database Error When Checking Target Registration", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Create test data
		sourceMerchant := &merchant.Merchant{
			UUID:       "source-merchant-id",
			ExternalId: "source-external-id",
			Name:       "Source Merchant",
		}

		targetMerchant := &merchant.Merchant{
			UUID:       "target-merchant-id",
			ExternalId: "target-external-id",
			Name:       "Target Merchant",
		}

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(sourceMerchant, nil)
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "target-merchant-id").Return(targetMerchant, nil)

		dbErr := errors.New("database error")
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "target-external-id").Return(nil, dbErr)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), dbErr.Error())
		mockMerchantRepo.AssertExpectations(t)
		mockQrisRepo.AssertExpectations(t)
	})

	t.Run("QR Registration Not Found", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Create test data
		sourceMerchant := &merchant.Merchant{
			UUID:       "source-merchant-id",
			ExternalId: "source-external-id",
			Name:       "Source Merchant",
		}

		targetMerchant := &merchant.Merchant{
			UUID:       "target-merchant-id",
			ExternalId: "target-external-id",
			Name:       "Target Merchant",
		}

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(sourceMerchant, nil)
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "target-merchant-id").Return(targetMerchant, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "target-external-id").Return(nil, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "source-external-id").Return(nil, nil)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), constant.ErrDataNotFound.Error())
		mockMerchantRepo.AssertExpectations(t)
		mockQrisRepo.AssertExpectations(t)
	})

	t.Run("QR Registration Status Not SUCCESS", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Create test data
		sourceMerchant := &merchant.Merchant{
			UUID:       "source-merchant-id",
			ExternalId: "source-external-id",
			Name:       "Source Merchant",
		}

		targetMerchant := &merchant.Merchant{
			UUID:       "target-merchant-id",
			ExternalId: "target-external-id",
			Name:       "Target Merchant",
		}

		sourceRegistration := &qrisModel.Registration{
			Id:                       "existing-reg-id",
			ExternalId:               "source-external-id",
			Acquirer:                 "BNC",
			MerchantType:             "Merchant",
			AcquirerParentMerchantId: "parent-merchant-id",
			MerchantName:             "Source Merchant",
			MerchantShortName:        "SM",
			Status:                   "PENDING", // Not SUCCESS
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(sourceMerchant, nil)
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "target-merchant-id").Return(targetMerchant, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "target-external-id").Return(nil, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "source-external-id").Return(sourceRegistration, nil)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), constant.ErrInvalidStatus.Error())
		mockMerchantRepo.AssertExpectations(t)
		mockQrisRepo.AssertExpectations(t)
	})

	t.Run("Database Error When Saving New Registration", func(t *testing.T) {
		// Setup mocks
		mockQrisRepo := mockRepo.NewIQrisRepository(t)
		mockMerchantRepo := mockRepo.NewIMerchantRepository(t)
		mockSnapRepo := mockRepo.NewISnapCoreRepository(t)

		// Create test data
		sourceMerchant := &merchant.Merchant{
			UUID:       "source-merchant-id",
			ExternalId: "source-external-id",
			Name:       "Source Merchant",
		}

		targetMerchant := &merchant.Merchant{
			UUID:       "target-merchant-id",
			ExternalId: "target-external-id",
			Name:       "Target Merchant",
		}

		callbackDetailRaw := types.NullJSONText{}
		_ = callbackDetailRaw.Scan([]byte(`{"applymentCode": "123", "resultCode": "SUCCESS"}`))

		acquirerMerchantId := "acquirer-merchant-id"

		sourceRegistration := &qrisModel.Registration{
			Id:                       "existing-reg-id",
			ExternalId:               "source-external-id",
			Acquirer:                 "BNC",
			MerchantType:             "Merchant",
			AcquirerParentMerchantId: "parent-merchant-id",
			MerchantName:             "Source Merchant",
			MerchantShortName:        "SM",
			Status:                   constant.QrRegistrationStatusSuccess,
			AcquirerMerchantId:       &acquirerMerchantId,
			CallbackDetailRaw:        callbackDetailRaw,
			CallbackDatetime:         sql.NullTime{Time: time.Now(), Valid: true},
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		// Set expectations
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "source-merchant-id").Return(sourceMerchant, nil)
		mockMerchantRepo.On("FindMerchantByID", mock.AnythingOfType("*context.valueCtx"), "target-merchant-id").Return(targetMerchant, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "target-external-id").Return(nil, nil)
		mockQrisRepo.On("FindQrRegistrationByExternalID", mock.AnythingOfType("*context.valueCtx"), "source-external-id").Return(sourceRegistration, nil)

		dbErr := errors.New("database error")
		mockQrisRepo.On("InitRegistration", mock.AnythingOfType("*context.valueCtx"), mock.MatchedBy(func(r interface{}) bool {
			return true // Match any registration object
		})).Return(dbErr)

		// Create service
		service := qrisService.New(mockLog, mockQrisRepo, mockMerchantRepo, mockSnapRepo)

		// Execute
		request := &qrisModel.DuplicateRegistrationReq{
			SourceMerchantId: "source-merchant-id",
			TargetMerchantId: "target-merchant-id",
		}
		id, err := service.DuplicateRegistration(ctx, request)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), dbErr.Error())
		mockMerchantRepo.AssertExpectations(t)
		mockQrisRepo.AssertExpectations(t)
	})
}
