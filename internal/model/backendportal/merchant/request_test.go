package merchant_test

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadDocumentReqToInsertData(t *testing.T) {
	req := &UploadDocumentReq{
		MerchantId: uuid.NewString(),
		Type:       "nationalIdentityCard",
		Identifier: "123456789",
		CreatedBy:  "John Wick",
	}
	docLocation := &DocLocation{
		Bucket: "test-bucket",
		Object: "/documents/merchants/xxxxxxxxxxxx.jpg",
	}
	buf, err := json.Marshal(docLocation)
	require.NoError(t, err)

	actual := req.ToInsertData(docLocation)

	require.NotEmpty(t, actual.Id)
	require.False(t, actual.CreatedAt.IsZero())
	require.False(t, actual.UpdatedAt.IsZero())
	require.False(t, actual.ApprovedAt.Time.IsZero())

	want := &Document{
		Id:         actual.Id,
		MerchantId: req.MerchantId,
		Type:       req.Type,
		Identifier: req.Identifier,
		Location:   buf,
		Status:     constant.StatusApproved,
		CreatedBy:  req.CreatedBy,
		CreatedAt:  actual.CreatedAt,
		ApprovedBy: req.CreatedBy,
		ApprovedAt: actual.ApprovedAt,
		UpdatedAt:  actual.UpdatedAt,
	}
	assert.Equal(t, want, actual)
}

func TestUpsertBoardOfDirectorReqToUpsertData(t *testing.T) {
	docLocation := &DocLocation{
		Bucket: "bucket",
		Object: "object.jpg",
	}
	bufDocLocation, err := json.Marshal(docLocation)
	require.NoError(t, err)

	dataPost := UpsertBoardOfDirectorReq{
		Method:         constant.ActionPost,
		MerchantId:     uuid.NewString(),
		Position:       "Director",
		Name:           "John Wick",
		IdentityNumber: "123456789",
		PositionLong:   "Director Exclusive",
		CreatedBy:      "Hendro",
	}
	actualPost := dataPost.ToUpsertData(docLocation)

	for _, val := range []string{actualPost.Id, actualPost.CreatedBy, actualPost.ApprovedBy} {
		require.NotEmpty(t, val)
	}
	for _, val := range []time.Time{actualPost.CreatedAt, actualPost.UpdatedAt, actualPost.ApprovedAt.Time} {
		require.False(t, val.IsZero())
	}

	wantPost := &BoardOfDirector{
		MerchantId:     dataPost.MerchantId,
		Position:       dataPost.Position,
		Name:           dataPost.Name,
		IdentityNumber: dataPost.IdentityNumber,
		IdentityFile:   bufDocLocation,
		PositionLong:   dataPost.PositionLong,
		Status:         constant.StatusApproved,
		CreatedBy:      dataPost.CreatedBy,
		ApprovedBy:     dataPost.CreatedBy,
		Id:             actualPost.Id,
		CreatedAt:      actualPost.CreatedAt,
		ApprovedAt:     actualPost.ApprovedAt,
		UpdatedAt:      actualPost.UpdatedAt,
	}
	assert.Equal(t, wantPost, actualPost)

	dataPut := UpsertBoardOfDirectorReq{
		Method:         constant.ActionPut,
		MerchantId:     uuid.NewString(),
		Position:       "Commissioner",
		Name:           "Endru",
		IdentityNumber: "987654321",
		PositionLong:   "Commissioner Idependent",
	}
	actualPut := dataPut.ToUpsertData(docLocation)

	for _, val := range []string{actualPut.Id, actualPut.CreatedBy, actualPut.ApprovedBy, actualPut.Status} {
		require.Empty(t, val)
	}
	for _, val := range []time.Time{actualPut.CreatedAt, actualPut.ApprovedAt.Time} {
		require.True(t, val.IsZero())
	}

	wantPut := &BoardOfDirector{
		MerchantId:     dataPut.MerchantId,
		Position:       dataPut.Position,
		Name:           dataPut.Name,
		IdentityNumber: dataPut.IdentityNumber,
		IdentityFile:   bufDocLocation,
		PositionLong:   dataPut.PositionLong,
		UpdatedAt:      actualPut.UpdatedAt,
	}
	assert.Equal(t, wantPut, actualPut)
}

func TestToOnBehalfFeeConfig(t *testing.T) {
	request := &CreateFeeConfigOnBehalfRequest{
		MerchantId:    "123456",
		Type:          "DIRECT",
		SubMerchantId: util.ValueToPtr("654321"),
		Reference:     "PAYMENT",
		PaymentMethod: util.ValueToPtr("VIRTUAL_ACCOUNT"),
		AmountType:    "AMOUNT",
		Amount:        10_000,
	}

	want := &OnBehalfFeeConfig{
		MerchantId:    "123456",
		Type:          "DIRECT",
		SubMerchantId: util.ValueToPtr("654321"),
		Reference:     "PAYMENT",
		PaymentMethod: util.ValueToPtr("VIRTUAL_ACCOUNT"),
		AmountType:    "AMOUNT",
		Amount:        10_000,
	}
	actual := request.ToOnBehalfFeeConfig()

	want.Id = actual.Id
	want.CreatedAt, want.UpdatedAt = actual.CreatedAt, actual.UpdatedAt

	assert.NotEmpty(t, actual.Id)
	assert.NotEmpty(t, actual.CreatedAt)
	assert.NotEmpty(t, actual.UpdatedAt)
	assert.Nil(t, actual.DeletedAt)
	assert.Equal(t, want, actual)
}

var asiaJakartaLoc, _ = time.LoadLocation(constant.TimeLoc)

func TestBillingDateRangeRequest(t *testing.T) {

	tests := []struct {
		request       BillingDateRangeRequest
		wantErr       error
		wantStartDate time.Time
		wantEndDate   time.Time
	}{
		{
			request: BillingDateRangeRequest{
				StrStartDate: "XXX",        // NOSONAR
				StrEndDate:   "2025-06-17", // NOSONAR
			},
			wantErr: &time.ParseError{
				Layout:     "2006-01-02", // NOSONAR
				Value:      "XXX",        // NOSONAR
				LayoutElem: "2006",       // NOSONAR
				ValueElem:  "XXX",        // NOSONAR
			},
		},
		{
			request: BillingDateRangeRequest{
				StrStartDate: "2025-06-17", // NOSONAR
				StrEndDate:   "YYY",        // NOSONAR
			},
			wantStartDate: time.Date(2025, 06, 17, 0, 0, 0, 0, asiaJakartaLoc), // NOSONAR
			wantErr: &time.ParseError{
				Layout:     "2006-01-02", // NOSONAR
				Value:      "YYY",        // NOSONAR
				LayoutElem: "2006",       // NOSONAR
				ValueElem:  "YYY",        // NOSONAR
			},
		},
		{
			request: BillingDateRangeRequest{
				StrStartDate: "2025-06-18", // NOSONAR
				StrEndDate:   "2025-06-17", // NOSONAR
			},
			wantErr:       constant.ErrInvalidDateRange,
			wantStartDate: time.Date(2025, 06, 17, 17, 0, 0, 0, time.UTC),
			wantEndDate:   time.Date(2025, 06, 17, 16, 59, 59, 999000000, time.UTC),
		},
		{
			request: BillingDateRangeRequest{
				StrStartDate: "2025-06-17", // NOSONAR
				StrEndDate:   "2025-06-17", // NOSONAR
			},
			wantStartDate: time.Date(2025, 06, 16, 17, 0, 0, 0, time.UTC),           // NOSONAR
			wantEndDate:   time.Date(2025, 06, 17, 16, 59, 59, 999000000, time.UTC), // NOSONAR
		},
		{
			request: BillingDateRangeRequest{
				StrStartDate: "2025-01-01", // NOSONAR
				StrEndDate:   "2025-02-28", // NOSONAR
			},
			wantStartDate: time.Date(2024, 12, 31, 17, 0, 0, 0, time.UTC),           // NOSONAR
			wantEndDate:   time.Date(2025, 02, 28, 16, 59, 59, 999000000, time.UTC), // NOSONAR
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Start: %s End: %s", test.request.StrStartDate, test.request.StrEndDate), func(t *testing.T) {
			assert.Equal(t, test.wantErr, test.request.ParseDateRangeRequestFromAsiaJakartaToUtc())
			assert.Equal(t, test.wantStartDate, test.request.StartDate)
			assert.Equal(t, test.wantEndDate, test.request.EndDate)
		})
	}
}

func TestUpsertBoardOfDirectorReqValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		request UpsertBoardOfDirectorReq
		wantErr error
	}{
		{
			name: "SUCCESS: Valid Director with identity (POST)",
			request: UpsertBoardOfDirectorReq{
				Method:         constant.ActionPost,
				Position:       constant.MerchantBODPositionDirector,
				Name:           "John Doe",
				IdentityNumber: "123456789",
				IdentityFile:   &multipart.FileHeader{Filename: "id.jpg"},
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Valid Shareholder with shares (POST)",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPost,
				Position: constant.MerchantBODPositionShareholder,
				Name:     "Jane Smith",
				Shares:   "25.5",
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Valid Commissioner (PUT)",
			request: UpsertBoardOfDirectorReq{
				Method:         constant.ActionPut,
				Position:       constant.MerchantBODPositionCommissioner,
				Name:           "Bob Johnson",
				IdentityNumber: "987654321",
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Valid Shareholder update with shares (PUT)",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPut,
				Position: constant.MerchantBODPositionShareholder,
				Name:     "Alice Brown",
				Shares:   "15.0",
			},
			wantErr: nil,
		},
		{
			name: "ERROR: Invalid position",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPost,
				Position: "InvalidPosition",
				Name:     "John Doe",
			},
			wantErr: constant.ErrInvalidMerchantBODPosition,
		},
		{
			name: "ERROR: Shareholder POST without shares",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPost,
				Position: constant.MerchantBODPositionShareholder,
				Name:     "Jane Smith",
			},
			wantErr: constant.ErrMerchantBODMandatoryShares,
		},
		{
			name: "ERROR: Shareholder POST with zero shares",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPost,
				Position: constant.MerchantBODPositionShareholder,
				Name:     "Jane Smith",
				Shares:   "",
			},
			wantErr: constant.ErrMerchantBODMandatoryShares,
		},
		{
			name: "ERROR: Shareholder POST with negative shares",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPost,
				Position: constant.MerchantBODPositionShareholder,
				Name:     "Jane Smith",
				Shares:   "-5.0",
			},
			wantErr: constant.ErrMerchantBODInvalidShares,
		},
		{
			name: "ERROR: Shareholder POST with 100+ shares",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPost,
				Position: constant.MerchantBODPositionShareholder,
				Name:     "Jane Smith",
				Shares:   "100.900",
			},
			wantErr: constant.ErrMerchantBODInvalidShares,
		},
		{
			name: "ERROR: Shareholder PUT with negative shares",
			request: UpsertBoardOfDirectorReq{
				Method:   constant.ActionPut,
				Position: constant.MerchantBODPositionShareholder,
				Name:     "Jane Smith",
				Shares:   "-10.0",
			},
			wantErr: constant.ErrMerchantBODInvalidShares,
		},
		{
			name: "ERROR: Director POST without identity file",
			request: UpsertBoardOfDirectorReq{
				Method:         constant.ActionPost,
				Position:       constant.MerchantBODPositionDirector,
				Name:           "John Doe",
				IdentityNumber: "123456789",
			},
			wantErr: constant.ErrMerchantBODMandatoryIdentity,
		},
		{
			name: "ERROR: Commissioner POST without identity number",
			request: UpsertBoardOfDirectorReq{
				Method:       constant.ActionPost,
				Position:     constant.MerchantBODPositionCommissioner,
				Name:         "Bob Johnson",
				IdentityFile: &multipart.FileHeader{Filename: "id.jpg"},
			},
			wantErr: constant.ErrMerchantBODMandatoryIdentity,
		},
		{
			name: "ERROR: President Director with negative shares",
			request: UpsertBoardOfDirectorReq{
				Method:         constant.ActionPost,
				Position:       constant.MerchantBODPositionPresidentDirector,
				Name:           "Mike Wilson",
				IdentityNumber: "555666777",
				IdentityFile:   &multipart.FileHeader{Filename: "id.jpg"},
				Shares:         "-1",
			},
			wantErr: constant.ErrMerchantBODInvalidShares,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.ValidateRequest()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
