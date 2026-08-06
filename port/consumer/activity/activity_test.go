package activity_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/activity"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/activity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestInsert(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	service := serviceMocks.NewIActivityService(t)

	consumer := New(logger, service)

	params := map[string]any{
		"Key1": "Value",
		"Key2": 1,
		"Key3": true,
	}
	rawParams, err := json.Marshal(params)
	require.NoError(t, err)

	body1 := &pb.Activity{
		Id:         uuid.NewString(),
		MerchantId: "123456",
	}
	body1.Parameter, err = anypb.New(wrapperspb.Bytes(rawParams))
	require.NoError(t, err)

	body2 := &pb.Activity{
		Id:         uuid.NewString(),
		MerchantId: "654321",
	}
	body2.Parameter, err = anypb.New(wrapperspb.String("Hello World"))
	require.NoError(t, err)

	body3 := &pb.Activity{
		Id:         uuid.NewString(),
		MerchantId: "1111",
		Parameter:  nil,
	}

	body4 := &pb.Activity{
		Id:         uuid.NewString(),
		MerchantId: "valid-company-id",
		Parameter:  nil,
	}

	rawBody1, _ := proto.Marshal(body1)
	rawBody2, _ := proto.Marshal(body2)
	rawBody3, _ := proto.Marshal(body3)
	rawBody4, _ := proto.Marshal(body4)

	errDuplicate := mysql.MySQLError{
		Number:  1062,
		Message: "duplicate entry",
	}

	tests := []struct {
		name      string
		body      []byte
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Invalid protobuf format",
			body: []byte(`{"message":"hello"}`),
			setupMock: func() {
			},
			wantErr: errors.New("unmarshal protobuf: cannot parse invalid wire-format data"),
		},
		{
			name: "ERROR:Invalid parameter wrappers type",
			body: rawBody2,
			setupMock: func() {
			},
			wantErr: errors.New(`unmarshal wrappers bytes type: mismatched message type`),
		},
		{
			name: "ERROR:Some error", // NOSONAR
			body: rawBody1,
			setupMock: func() {
				service.On("Create", c.ValueCtxMockType(), mock.AnythingOfType("*activityModel.Activity")).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Non nil parameters",
			body: rawBody1,
			setupMock: func() {
				service.On("Create", c.ValueCtxMockType(), mock.AnythingOfType("*activityModel.Activity")).Return(nil).Once()
			},
		},
		{
			name: "SUCCESS:Nil parameter",
			body: rawBody3,
			setupMock: func() {
				service.On("Create", c.ValueCtxMockType(), mock.AnythingOfType("*activityModel.Activity")).Return(nil).Once()
			},
		},
		{
			name: "should not return error when create activity error duplicate",
			body: rawBody4,
			setupMock: func() {
				service.On("Create", c.ValueCtxMockType(), mock.AnythingOfType("*activityModel.Activity")).Return(&errDuplicate).Once()
			},
		},
		{
			name: "should not return error when create activity error no affected",
			body: rawBody4,
			setupMock: func() {
				service.On("Create", c.ValueCtxMockType(), mock.AnythingOfType("*activityModel.Activity")).Return(c.ErrNoRowsAffected).Once()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			assert.Equal(t, test.wantErr, consumer.Insert(context.Background(), test.body, ""))
		})
	}
}
