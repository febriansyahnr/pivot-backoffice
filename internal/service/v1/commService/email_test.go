package commService_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/paperCommunication"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/commService"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
)

func TestPostEmailService(t *testing.T) {
	paperComm := repoMocks.NewIPaperCommunicationRepository(t)

	service := New(paperComm)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				paperComm.On(
					"SendEmailV1", c.ValueCtxMockType(), c.StringMockType(), c.PtrPaperCommEmailRequestMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				paperComm.On(
					"SendEmailV1", c.ValueCtxMockType(), c.StringMockType(), c.PtrPaperCommEmailRequestMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, service.PostEmailService(context.Background(), "", &paperCommunication.Email{}))
		})
	}
}
