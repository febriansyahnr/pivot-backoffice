package withdrawalRepository_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/withdrawal"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetadataById(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.JSONTextMockType(), c.TimeMockType(), "123",
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.JSONTextMockType(), c.TimeMockType(), "123",
				).Once().Return(false, nil)
			},
			wantErr: c.ErrNoRowsAffected,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.JSONTextMockType(), c.TimeMockType(), "123",
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.UpdateMetadataById(context.Background(), "123", &withdrawal.Metadata{}))
		})
	}
}
