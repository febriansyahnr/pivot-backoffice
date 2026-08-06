package qris_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/qris"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	pdfMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/pdf"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	gcsModel "github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegistrationX(t *testing.T) {
	pdf := pdfMock.NewPDFGenerator(t)
	pdf.On(
		"MergeFilesToPDF", c.ValueCtxMockType(), mock.AnythingOfType("[]pdf.MergeFile"),
	).Return([]byte("Hello"), nil)

	gcs := gcsMock.NewGCSService(t)
	gcs.On(
		"UploadFile", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*bytes.Reader"), c.BoolMockType(),
	).Return(&gcsModel.UploadMultipart{}, nil)
	gcs.On("ReadAll", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).Return([]byte(`{}`), nil)

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	merchantRepo := repoMocks.NewIMerchantRepository(t)

	regId := uuid.NewString()
	traceId := uuid.NewString()
	merchantInfo := &merchant.QrisMerchant{
		RegId:      regId,
		MID:        "0001",
		ExternalId: uuid.NewString(),
		Name:       "Harsya",
		ShortName:  "HSY",
		MCC:        "5999",
		Address: qris.Address{
			Province: 1,
			City:     2,
			District: 3,
			PostCode: "4",
			Detail:   "5",
		},
		Documents: []merchant.QrisDocument{
			{Type: "NationalIdentityCard"},
			{Type: "BusinessLicense"},
			{Type: "TaxIdentification"},
			{Type: "BusinessRegistration"},
			{Type: "CertificateIncorporation"},
			{Type: "CertificateNo40"},
			{Type: "CertificateLastAmendment"},
			{Type: "CertificateDeedAmendment"},
			{Type: "CertificateAmendmentAct"},
			{Type: "CertificateEstablishment"},
			{Type: "CertificateTaxRegistration"},
			{Type: "BusinessEnvironmentPhoto"},
		},
		BODCount: 1, BOCCount: 1, BoardOfDirectors: []merchant.QrisBOD{{Position: c.PositionCommissioner}},
		RegStatus: c.FillingFormReg,
	}
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	successMocks := []func(p1 *gcsMock.GCSService, p2 *repoMocks.ISnapCoreRepository, p3 *repoMocks.IQrisRepository){}

	tests := []struct {
		name       string
		setupMock  func(gcs *gcsMock.GCSService, snapCore *repoMocks.ISnapCoreRepository, repo *repoMocks.IQrisRepository)
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:Find merchant for QRIS registration",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: c.ErrMerchantIDNotValid.Error(),
		},
		{
			name: "ERROR:Registration has been submitted",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchant.QrisMerchant{RegStatus: c.SubmittedReg}, nil)
			},
			wantErr: "registration has been submitted",
		},
		{
			name: "ERROR:Merchant is already registered",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchant.QrisMerchant{RegStatus: c.SuccessReg}, nil)
			},
			wantErr: "merchant is already registered",
		},
		{
			name: "ERROR:Incomplete data",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				tmp := *merchantInfo
				tmp.ExternalId = ""

				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&tmp, nil)
			},
			wantErr: `external_id is a required field`,
		},
		{
			name: "ERROR:Document is incomplete",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				tmp := *merchantInfo
				tmp.Documents = nil

				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&tmp, nil)
			},
			wantErr: `document required`,
		},
		{
			name: "ERROR:BOD data is empty",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				tmp := *merchantInfo
				tmp.BODCount = 0
				tmp.BOCCount = 1
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&tmp, nil)
			},
			wantErr: `{"bod":"board of directors is required"}`,
		},
		{
			name: "ERROR:BOC data is empty",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				tmp := *merchantInfo
				tmp.BODCount = 1
				tmp.BOCCount = 0
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&tmp, nil)
			},
			wantErr: `{"boc":"board of commissioner is required"}`,
		},
		{
			name: "ERROR:MCC is required",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
				tmp := *merchantInfo
				tmp.MCC = ""
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&tmp, nil)
			},
			wantErr: "mcc is required",
		},
		{
			name: "ERROR:Init registration",
			setupMock: func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, repo *repoMocks.IQrisRepository) {
				tmp := *merchantInfo
				tmp.RegStatus = ""
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&tmp, nil)

				repo.On(
					"InitRegistration", c.ValueCtxMockType(), mock.AnythingOfType("*qris.Registration"),
				).Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:QRIS update document",
			setupMock: func(gcs *gcsMock.GCSService, snapCore *repoMocks.ISnapCoreRepository, repo *repoMocks.IQrisRepository) {
				merchantRepo.On(
					"FindMerchantForQrRegistration", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(merchantInfo, nil)

				snapCore.On("QrUploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.UploadDocumentReq")).Return(nil, c.ErrSomeErrorForUnitTest)
				successMocks = append(successMocks, func(_ *gcsMock.GCSService, p2 *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
					p2.On("QrUploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.UploadDocumentReq")).Return(&snapCoreModel.UploadDocumentResp{}, nil)
				})
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:Update uploaded document",
			setupMock: func(gcs *gcsMock.GCSService, snapCore *repoMocks.ISnapCoreRepository, repo *repoMocks.IQrisRepository) {
				for _, f := range successMocks {
					f(gcs, snapCore, repo)
				}
				repo.On(
					"UpdateUploadedDocument", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.UpdateDocument"),
				).Return(c.ErrSomeErrorForUnitTest)
				successMocks = append(successMocks, func(_ *gcsMock.GCSService, _ *repoMocks.ISnapCoreRepository, p3 *repoMocks.IQrisRepository) {
					p3.On(
						"UpdateUploadedDocument", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.UpdateDocument"),
					).Return(nil)
				})
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:QRIS final registration",
			setupMock: func(gcs *gcsMock.GCSService, snapCore *repoMocks.ISnapCoreRepository, repo *repoMocks.IQrisRepository) {
				for _, f := range successMocks {
					f(gcs, snapCore, repo)
				}
				snapCore.On(
					"QrFinalRegistration", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.RegistrationReq"),
				).Return(c.ErrSomeErrorForUnitTest)
				successMocks = append(successMocks, func(_ *gcsMock.GCSService, p2 *repoMocks.ISnapCoreRepository, _ *repoMocks.IQrisRepository) {
					p2.On(
						"QrFinalRegistration", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.RegistrationReq"),
					).Return(nil)
				})
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:Update registration status",
			setupMock: func(gcs *gcsMock.GCSService, snapCore *repoMocks.ISnapCoreRepository, repo *repoMocks.IQrisRepository) {
				for _, f := range successMocks {
					f(gcs, snapCore, repo)
				}
				repo.On(
					"UpdateRegistrationStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "SUCCESS",
			setupMock: func(gcs *gcsMock.GCSService, snapCore *repoMocks.ISnapCoreRepository, repo *repoMocks.IQrisRepository) {
				for _, f := range successMocks {
					f(gcs, snapCore, repo)
				}
				repo.On(
					"UpdateRegistrationStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil)
			},
			wantResult: regId,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repoMocks.NewIQrisRepository(t)
			snapRepo := repoMocks.NewISnapCoreRepository(t)
			service := New(logger, repo, merchantRepo, snapRepo, WithGCSService(gcs), WithPDFGenerator(pdf), WithServiceConfig(&config.Config{}))

			test.setupMock(gcs, snapRepo, repo)

			acquirer := ""
			if test.name == "ERROR:MCC is required" {
				acquirer = constant.BANK_ACQUIRER_BNC
			}

			if id, err := service.Registration(ctx, &qris.RegistrationReq{Acquirer: acquirer}); test.wantErr == "" {
				require.NoError(t, err)
				assert.NotEmpty(t, id)
				assert.Equal(t, test.wantResult, id)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestRegistrationCallback(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIQrisRepository(t)

	data := &qris.Registration{
		Id:     uuid.NewString(),
		Status: c.SubmittedReg,
	}
	service := New(logger, repo, nil, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:" + c.ErrSomeErrorForUnitTest.Error(),
			setupMock: func() {
				repo.On(
					"FindQrRegistrationForValidationById", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "WARN:data not found",
			setupMock: func() {
				repo.On("FindQrRegistrationForValidationById", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, nil)
			},
		},
		{
			name: "WARN:status already success",
			setupMock: func() {
				repo.On(
					"FindQrRegistrationForValidationById", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&qris.Registration{Status: c.SuccessReg}, nil)
			},
		},
		{
			name: "ERROR:update callback registration",
			setupMock: func() {
				repo.On(
					"FindQrRegistrationForValidationById", c.ValueCtxMockType(), c.StringMockType(),
				).Return(data, nil)
				repo.On(
					"UpdateCallbackQrRegistration", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.RegistrationCallback"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On(
					"UpdateCallbackQrRegistration", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.RegistrationCallback"),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if err := service.RegistrationCallback(context.Background(), &qris.RegistrationCallback{}); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestUpdateQrRegistration_Service(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	tests := []struct {
		name      string
		setupMock func(repo *repoMocks.IQrisRepository)
		wantErr   bool
	}{
		{
			name: "ERROR: repository returns error",
			setupMock: func(repo *repoMocks.IQrisRepository) {
				repo.On(
					"UpdateQrRegistration", mock.Anything, "reg-id", "merchant-id", "terminal-id",
				).Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: repository returns nil",
			setupMock: func(repo *repoMocks.IQrisRepository) {
				repo.On(
					"UpdateQrRegistration", mock.Anything, "reg-id", "merchant-id", "terminal-id",
				).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repoMocks.NewIQrisRepository(t)
			service := New(logger, repo, nil, nil)
			test.setupMock(repo)
			err := service.UpdateQrRegistration(context.Background(), "reg-id", "merchant-id", "terminal-id")
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
