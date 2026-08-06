package merchantRcn

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cimbProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantRcn"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	repoMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRcnService_FindByIDAndMerchantID(t *testing.T) {
	expectedMerchantRcn := merchantRcn.MerchantRcn{
		ID:                uuid.New(),
		MerchantID:        uuid.New(),
		PrincipalIssuer:   "CIMB_NIAGA",
		RealCardNumber:    base64.StdEncoding.EncodeToString([]byte("encrypted-card-number")),
		EncryptKMSVersion: "1",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockDecryptedCard := "1234567890123456"

	type fields struct {
		repo          repository.IMerchantRcnRepository
		cimbProcessor repository.ICimbProcessorRepository
		gcsEncryption encryption.GCSClient
	}
	type args struct {
		ctx        context.Context
		id         string
		merchantId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *merchantRcn.MerchantRcnResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			fields: fields{
				repo: func() repository.IMerchantRcnRepository {
					mockRepo := new(repoMock.IMerchantRcnRepository)
					mockRepo.On("FindByIDAndMerchantID", mock.Anything, mock.Anything, mock.Anything).Return(&expectedMerchantRcn, nil)
					return mockRepo
				}(),
				cimbProcessor: func() repository.ICimbProcessorRepository {
					mockCimb := new(repoMock.ICimbProcessorRepository)
					mockCimb.On("InquiryCorporateCreditCard", mock.Anything, mock.Anything).Return(&cimbProcessorModel.InquiryCorporateCreditCardResponse{
						PartnerReferenceNo: "PRN123456",
						AdditionalInfo: cimbProcessorModel.AdditionalInfoResponse{
							ReferenceNo: "REF123456",
							Data: cimbProcessorModel.Data{
								CardInformation: []cimbProcessorModel.CardInfo{
									{
										BankCardNo: "1234567890123456",
										FullName:   "John Doe",
									},
								},
							},
						},
					}, nil)
					return mockCimb
				}(),
				gcsEncryption: func() encryption.GCSClient {
					mockGCS := new(mocks.GCSClient)
					mockGCS.On("DecryptSymmetric", mock.Anything, mock.Anything).Return(mockDecryptedCard, nil)
					return mockGCS
				}(),
			},
			args: args{
				ctx:        context.Background(),
				id:         "uuid-uuid-uuid",
				merchantId: "merchant-id",
			},
			want: &merchantRcn.MerchantRcnResponse{
				PartnerReferenceNo: "PRN123456",
				AdditionalInfo: merchantRcn.AdditionalInfo{
					ReferenceNo: "REF123456",
					CardInformation: []merchantRcn.CardInfo{
						{
							BankCardNo: "123456******3456",
							FullName:   "J**n D*e",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "MERCHANT NOT FOUND",
			fields: fields{
				repo: func() repository.IMerchantRcnRepository {
					mockRepo := new(repoMock.IMerchantRcnRepository)
					mockRepo.On("FindByIDAndMerchantID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
					return mockRepo
				}(),
			},
			args: args{
				ctx:        context.Background(),
				id:         "invalid-id",
				merchantId: "invalid-merchant-id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "REPOSITORY ERROR",
			fields: fields{
				repo: func() repository.IMerchantRcnRepository {
					mockRepo := new(repoMock.IMerchantRcnRepository)
					mockRepo.On("FindByIDAndMerchantID", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
					return mockRepo
				}(),
			},
			args: args{
				ctx:        context.Background(),
				id:         "uuid-uuid-uuid",
				merchantId: "merchant-id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Base 64 FAILURE",
			fields: fields{
				repo: func() repository.IMerchantRcnRepository {
					mockRepo := new(repoMock.IMerchantRcnRepository)
					mockRepo.On("FindByIDAndMerchantID", mock.Anything, mock.Anything, mock.Anything).Return(&merchantRcn.MerchantRcn{
						ID:                uuid.New(),
						MerchantID:        uuid.New(),
						PrincipalIssuer:   "CIMB_NIAGA",
						RealCardNumber:    "encrypted-card-number",
						EncryptKMSVersion: "1",
						CreatedAt:         time.Now(),
						UpdatedAt:         time.Now(),
					}, nil)
					return mockRepo
				}(),
			},
			args: args{
				ctx:        context.Background(),
				id:         "uuid-uuid-uuid",
				merchantId: "merchant-id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "DECRYPTION FAILURE",
			fields: fields{
				repo: func() repository.IMerchantRcnRepository {
					mockRepo := new(repoMock.IMerchantRcnRepository)
					mockRepo.On("FindByIDAndMerchantID", mock.Anything, mock.Anything, mock.Anything).Return(&expectedMerchantRcn, nil)
					return mockRepo
				}(),
				gcsEncryption: func() encryption.GCSClient {
					mockGCS := new(mocks.GCSClient)
					mockGCS.On("DecryptSymmetric", mock.Anything, mock.Anything).Return("", errors.New("decryption error"))
					return mockGCS
				}(),
			},
			args: args{
				ctx:        context.Background(),
				id:         "uuid-uuid-uuid",
				merchantId: "merchant-id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "INQUIRY API FAILURE",
			fields: fields{
				repo: func() repository.IMerchantRcnRepository {
					mockRepo := new(repoMock.IMerchantRcnRepository)
					mockRepo.On("FindByIDAndMerchantID", mock.Anything, mock.Anything, mock.Anything).Return(&expectedMerchantRcn, nil)
					return mockRepo
				}(),
				cimbProcessor: func() repository.ICimbProcessorRepository {
					mockCimb := new(repoMock.ICimbProcessorRepository)
					mockCimb.On("InquiryCorporateCreditCard", mock.Anything, mock.Anything).Return(nil, errors.New("API error"))
					return mockCimb
				}(),
				gcsEncryption: func() encryption.GCSClient {
					mockGCS := new(mocks.GCSClient)
					mockGCS.On("DecryptSymmetric", mock.Anything, mock.Anything).Return(mockDecryptedCard, nil)
					return mockGCS
				}(),
			},
			args: args{
				ctx:        context.Background(),
				id:         "uuid-uuid-uuid",
				merchantId: "merchant-id",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MerchantRcnService{
				repo:          tt.fields.repo,
				cimbProcessor: tt.fields.cimbProcessor,
				gcsEncryption: tt.fields.gcsEncryption,
				logger:        mockLogger,
			}
			got, err := m.FindByIDAndMerchantID(tt.args.ctx, tt.args.id, tt.args.merchantId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByIDAndMerchantID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindByIDAndMerchantID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
