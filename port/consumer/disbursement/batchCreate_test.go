package disbursementConsumerController_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/disbursement"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/disbursement"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBatchCreateDisbursement(t *testing.T) {

	logger := loggerMock.NewILogger(t)
	service := serviceMocks.NewIDisbursementService(t)

	handler := New(logger, service, nil)

	logger.On(
		"Info", c.ValueCtxMockType(), "Incoming message on BatchCreateDisbursement", mock.Anything, mock.Anything,
	).Return()

	payload := &pb.BatchCreateDisbursementRequest{
		BulkId:       "REF001",
		MerchantId:   "1234",
		MerchantName: "Dummy",
		CreatedBy:    "Dashboard",
		TotalTrx:     1,
		AutoApprove:  true,
		Data:         []*pb.CreateSingleRequest{{}},
	}
	rawPayload, err := proto.Marshal(payload)
	require.NoError(t, err)

	requestMockType := mock.AnythingOfType("*disbursementModel.BatchCreateDisbursementRequest")
	tests := []struct {
		name      string
		body      []byte
		setupMock func()
		wantErr   error
	}{
		{
			name:    "ERROR:Unmarshal proto",
			body:    []byte(`BLABLA`),
			wantErr: c.ErrUnmarshalProto,
		},
		{
			name: "ERROR:Some error", // NOSONAR
			body: rawPayload,
			setupMock: func() {
				service.On(
					"BatchCreateDisbursement", c.ValueCtxMockType(), requestMockType,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			body: rawPayload,
			setupMock: func() {
				service.On("BatchCreateDisbursement", c.ValueCtxMockType(), requestMockType).Return(nil).Once()
			},
		},
		{
			name: "when got error of beneficiary limit, then should not got error",
			body: rawPayload,
			setupMock: func() {
				service.On("BatchCreateDisbursement", c.ValueCtxMockType(), requestMockType).Return(
					pkgErrs.New(response.HttpErrDailyLimitReached, c.ErrBeneficiaryLimitRestrictions),
				).Once()
			},
		},
		{
			name: "when got error of list beneficiary limit, then should not got error",
			body: rawPayload,
			setupMock: func() {
				service.On("BatchCreateDisbursement", c.ValueCtxMockType(), requestMockType).Return(
					&disbursementModel.ApprovalResultErr{},
				).Once()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			assert.Equal(t, test.wantErr, handler.BatchCreateDisbursement(context.Background(), test.body, ""))
		})
	}
}
