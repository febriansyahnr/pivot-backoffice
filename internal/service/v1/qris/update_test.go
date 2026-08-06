package qris_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
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

func TestReuploadDocument(t *testing.T) {
	gcs := gcsMock.NewGCSService(t)
	pdf := pdfMock.NewPDFGenerator(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIQrisRepository(t)
	snapCore := repoMocks.NewISnapCoreRepository(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	cfg := &config.Config{}
	service := New(logger, repo, merchantRepo, snapCore, WithGCSService(gcs), WithPDFGenerator(pdf), WithServiceConfig(cfg))

	gcs.On(
		"ReadAll", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
	).Return([]byte{}, nil)
	snapCore.On(
		"QrUploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.UploadDocumentReq"),
	).Return(&snapCoreModel.UploadDocumentResp{}, nil)

	traceId := uuid.NewString()
	merchantId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	const certificateEstablishmentType = "CertificateEstablishment"

	tests := []struct {
		name       string
		request    *qris.ReuploadDocumentReq
		setupMock  func()
		wantErr    string
		wantResult *qris.ReuploadDocumentResp
	}{
		{
			name: "ERROR:Find registration by id",
			setupMock: func() {
				repo.On(
					"FindRegistrationById", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("RM: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Registration not found",
			setupMock: func() {
				repo.On("FindRegistrationById", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, nil)
			},
			wantErr: "registration not found",
		},
		{
			name: "ERROR:Must be failed status",
			setupMock: func() {
				repo.On(
					"FindRegistrationById", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&qris.RegistrationMerchant{Status: c.SubmittedReg}, nil)
			},
			wantErr: "registration status must be failed to be able to reupload document",
		},
		{
			name: "ERROR:Find document by type",
			setupMock: func() {
				repo.On(
					"FindRegistrationById", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&qris.RegistrationMerchant{MerchantId: merchantId, Status: c.FailedReg}, nil)

				merchantRepo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("FD: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Get list merchant bod",
			request: &qris.ReuploadDocumentReq{
				DocumentType: certificateEstablishmentType,
			},
			setupMock: func() {
				merchantRepo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&merchant.Document{
					MerchantId: merchantId, Identifier: "-", ObjLocation: merchant.DocLocation{},
				}, nil)

				merchantRepo.On(
					"GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("MBOD: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Merging many document to single file pdf",
			request: &qris.ReuploadDocumentReq{
				DocumentType: certificateEstablishmentType,
			},
			setupMock: func() {
				merchantRepo.On(
					"GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType(),
				).Return([]merchant.BoardOfDirectorResp{
					{File: []byte(`{"bucket":"myBucket","object":"myObject"}`)},
				}, nil)

				pdf.On(
					"MergeFilesToPDF", c.ValueCtxMockType(), mock.AnythingOfType("[]pdf.MergeFile"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("MF: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Upload to GCS",
			request: &qris.ReuploadDocumentReq{
				DocumentType: certificateEstablishmentType,
			},
			setupMock: func() {
				pdf.On("MergeFilesToPDF", c.ValueCtxMockType(), mock.AnythingOfType("[]pdf.MergeFile")).Return([]byte{}, nil)

				gcs.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*bytes.Reader"), c.BoolMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("UP: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Update uploaded document",
			request: &qris.ReuploadDocumentReq{
				DocumentType: certificateEstablishmentType,
			},
			setupMock: func() {
				gcs.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*bytes.Reader"), c.BoolMockType(),
				).Return(&gcsModel.UploadMultipart{}, nil)

				repo.On(
					"UpdateUploadedDocument", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.UpdateDocument"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On("UpdateUploadedDocument", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.UpdateDocument")).Return(nil)
			},
			wantResult: &qris.ReuploadDocumentResp{Uploaded: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if test.request == nil {
				test.request = &qris.ReuploadDocumentReq{}
			}

			if resp, err := service.ReuploadDocument(ctx, test.request); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
