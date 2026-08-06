package transfer

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestToTransferResponse(t *testing.T) {
	var (
		now = time.Now().UTC()
	)

	transfer := &Transfer{
		UUID:              uuid.Max,
		MerchantID:        uuid.Max,
		RecipientID:       uuid.Max,
		ReferenceID:       "reference-id",
		TransferType:      constant.MoneyFlowDirect,
		Currency:          constant.CurrencyIDR,
		Amount:            10,
		Status:            constant.TransferStatusPending,
		Remarks:           "remarks",
		ReasonDescription: "description",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	resp := transfer.ToTransferResponse()
	assert.Equal(t, transfer.UUID.String(), resp.UUID)
	assert.Equal(t, transfer.RecipientID.String(), resp.RecipientID)
	assert.Equal(t, transfer.ReferenceID, resp.ReferenceID)
	assert.Equal(t, transfer.TransferType, resp.TransferType)
	assert.Equal(t, transfer.Amount, resp.Amount)
	assert.Equal(t, transfer.Status, resp.Status)
	assert.Equal(t, transfer.Remarks, resp.Remarks)
	assert.Equal(t, transfer.CreatedAt, resp.CreatedAt)
	assert.Equal(t, transfer.UpdatedAt, resp.UpdatedAt)

}

func TestValidate(t *testing.T) {

	testCases := []struct {
		Name    string
		Input   *TransferRequest
		WantErr bool
		Err     error
	}{
		{
			Name: "SUCCESS: Valid",
			Input: &TransferRequest{
				SourceMerchantID: uuid.New(),
				RecipientID:      uuid.New().String(),
				ParentMerchantID: uuid.New(),
				TransferType:     constant.MoneyFlowDirect,
				Amount:           1000,
				Remarks:          "Description test",
			},
			WantErr: false,
			Err:     nil,
		},
		{
			Name: "SUCCESS: Valid Transfer type",
			Input: &TransferRequest{
				SourceMerchantID: uuid.New(),
				RecipientID:      uuid.New().String(),
				ParentMerchantID: uuid.New(),
				TransferType:     constant.MoneyFlowIndirect,
				Amount:           1000,
				Remarks:          "Description test",
			},
			WantErr: false,
			Err:     nil,
		},
		{
			Name: "ERROR: Invalid participants",
			Input: &TransferRequest{
				TransferType: constant.MoneyFlowDirect,
				Amount:       10,
			},
			WantErr: true,
			Err:     constant.ErrInvalidMerchantId,
		},
		{
			Name: "ERROR: Same Recipient",
			Input: &TransferRequest{
				SourceMerchantID: uuid.Max,
				RecipientID:      uuid.Max.String(),
				ParentMerchantID: uuid.New(),
				TransferType:     constant.MoneyFlowDirect,
				Amount:           10,
			},
			WantErr: true,
			Err:     constant.ErrSameMerchant,
		},
		{
			Name: "ERROR: Invalid Transfer Type",
			Input: &TransferRequest{
				SourceMerchantID: uuid.Max,
				RecipientID:      uuid.New().String(),
				ParentMerchantID: uuid.New(),
				TransferType:     "",
				Amount:           10,
			},
			WantErr: true,
			Err:     constant.ErrInvalidTransferType,
		},
		{
			Name: "ERROR: Invalid Transfer Type With Non-Empty Value",
			Input: &TransferRequest{
				SourceMerchantID: uuid.New(),
				RecipientID:      uuid.New().String(),
				ParentMerchantID: uuid.New(),
				TransferType:     "INVALID_TYPE",
				Amount:           10,
			},
			WantErr: true,
			Err:     constant.ErrInvalidTransferType,
		},
		{
			Name: "ERROR: Invalid Amount",
			Input: &TransferRequest{
				SourceMerchantID: uuid.New(),
				RecipientID:      uuid.New().String(),
				ParentMerchantID: uuid.New(),
				TransferType:     constant.MoneyFlowDirect,
				Amount:           0,
			},
			WantErr: true,
			Err:     constant.ErrInvalidAmount,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Input.Validate()
			if tc.WantErr {
				if !assert.NotNil(t, err) {
					t.Errorf("Validate() error = %v, wantErr %v", err, tc.WantErr)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNewTransfer(t *testing.T) {

	var (
		sourceMerchantID = uuid.New()
		parentMerchantID = uuid.New()
		recipientID      = uuid.New()
	)

	testCases := []struct {
		Name     string
		Input    *TransferRequest
		Expected *Transfer
		WantErr  bool
		Err      error
	}{
		{
			Name: "SUCCESS: Valid",
			Input: &TransferRequest{
				SourceMerchantID: sourceMerchantID,
				RecipientID:      recipientID.String(),
				ReferenceID:      "reference-id",
				ParentMerchantID: parentMerchantID,
				TransferType:     constant.MoneyFlowDirect,
				Amount:           1000,
				Remarks:          "Description test",
			},
			Expected: &Transfer{
				MerchantID:   sourceMerchantID,
				RecipientID:  recipientID,
				ReferenceID:  "reference-id",
				TransferType: constant.MoneyFlowDirect,
				Currency:     constant.CurrencyIDR,
				Amount:       1000,
				Status:       constant.TransferStatusPending,
				Remarks:      "Description test",
			},
			WantErr: false,
			Err:     nil,
		},
		{
			Name: "ERROR: Failed Validation",
			Input: &TransferRequest{
				SourceMerchantID: recipientID,
				RecipientID:      recipientID.String(),
				ReferenceID:      "reference-id",
				ParentMerchantID: parentMerchantID,
				TransferType:     constant.MoneyFlowDirect,
				Amount:           1000,
				Remarks:          "Description test",
			},
			Expected: nil,
			WantErr:  true,
			Err:      constant.ErrSameMerchant,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			output, err := NewTransfer(tc.Input)
			if tc.WantErr {
				if !assert.NotNil(t, err) {
					t.Errorf("NewTransfer() error = %v, wantErr %v", err, tc.WantErr)
				}
			} else {
				assert.Nil(t, err)
				assert.NotEqual(t, uuid.Nil, output.UUID)
				assert.Equal(t, tc.Expected.MerchantID, output.MerchantID)
				assert.Equal(t, tc.Expected.RecipientID, output.RecipientID)
				assert.Equal(t, tc.Expected.ReferenceID, output.ReferenceID)
				assert.Equal(t, tc.Expected.TransferType, output.TransferType)
				assert.Equal(t, tc.Expected.Currency, output.Currency)
				assert.Equal(t, tc.Expected.Amount, output.Amount)
				assert.Equal(t, tc.Expected.Status, output.Status)
				assert.Equal(t, tc.Expected.Remarks, output.Remarks)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	var (
		now = time.Now().UTC()
	)

	transfer := &Transfer{
		UUID:              uuid.Max,
		MerchantID:        uuid.Max,
		RecipientID:       uuid.Max,
		ReferenceID:       "reference-id",
		TransferType:      constant.MoneyFlowDirect,
		Currency:          constant.CurrencyIDR,
		Amount:            10,
		Status:            constant.TransferStatusPending,
		Remarks:           "remarks",
		ReasonDescription: "",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	transfer.Update(constant.TransferStatusSuccess, "success")
	assert.Equal(t, constant.TransferStatusSuccess, transfer.Status)
	assert.Equal(t, "success", transfer.ReasonDescription)
}

func TestValidateAndAdjust(t *testing.T) {
	testCases := []struct {
		name    string
		input   *GetTransferListRequest
		wantErr bool
	}{
		{
			name: "SUCCESS: Transfer IN Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
			},
		},
		{
			name: "SUCCESS: Transfer IN Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				Type:       constant.TransferTypeIN,
			},
		},
		{
			name: "SUCCESS: Transfer OUT Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				Type:       constant.TransferTypeOUT,
			},
		},
		{
			name: "SUCCESS: Transfer Status Success Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				Status:     constant.TransferStatusSuccess,
			},
		},
		{
			name: "SUCCESS: Transfer Status Pending Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				Status:     constant.TransferStatusPending,
			},
		},
		{
			name: "SUCCESS: Transfer Status Failed Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				Status:     constant.TransferStatusFailed,
			},
		},
		{
			name: "SUCCESS: Transfer Sort by Amount Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				SortBy:     SortColAmount,
			},
		},
		{
			name: "SUCCESS: Transfer with StartDate Now Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				SortBy:     SortColAmount,
				StartDate:  time.Now().UTC(),
			},
		},
		{
			name: "SUCCESS: Transfer with EndDate Now Request",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				SortBy:     SortColAmount,
				EndDate:    time.Now().UTC(),
			},
		},
		{
			name: "ERROR: Empty merchant ID",
			input: &GetTransferListRequest{
				SortBy: SortColAmount,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid type",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				Type:       "MIXED",
				SortBy:     SortColAmount,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid status",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				Status:     "Unknown",
				SortBy:     SortColAmount,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid sort column",
			input: &GetTransferListRequest{
				MerchantID: uuid.New().String(),
				SortBy:     "unknown",
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.ValidateAndAdjust()
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestToListTransferResponse(t *testing.T) {
	merchantSenderID := uuid.New()
	merchantRecipientID := uuid.New()
	participantMap := map[string]string{
		merchantSenderID.String():    "merchant 1",
		merchantRecipientID.String(): "merchant 2",
	}

	testCases := []struct {
		name              string
		input             *Transfer
		requesterMerchant string
		expected          *ListTransferResponse
	}{
		{
			name: "SUCCESS: TRANSFER_IN Type",
			input: &Transfer{
				MerchantID:   merchantSenderID,
				RecipientID:  merchantRecipientID,
				TransferType: constant.MoneyFlowDirect,
				Amount:       100,
				ReferenceID:  "ref-id",
				Status:       constant.TransferStatusSuccess,
			},
			requesterMerchant: merchantSenderID.String(),
			expected: &ListTransferResponse{
				UUID:          "00000000-0000-0000-0000-000000000000",
				ReferenceID:   "ref-id",
				Type:          constant.TransferTypeIN,
				SenderID:      merchantSenderID.String(),
				SenderName:    "merchant 1",
				RecipientID:   merchantRecipientID.String(),
				RecipientName: "merchant 2",
				TransferType:  constant.MoneyFlowDirect,
				Amount:        100,
				Status:        constant.TransferStatusSuccess,
				Remarks:       "",
				UpdatedAt:     time.Time{},
				CreatedAt:     time.Time{},
			},
		},
		{
			name: "SUCCESS: TRANSFER_OUT Type",
			input: &Transfer{
				MerchantID:   merchantSenderID,
				RecipientID:  merchantRecipientID,
				TransferType: constant.MoneyFlowDirect,
				Direction:    constant.TransferTypeOUT,
				Amount:       100,
				ReferenceID:  "ref-id",
				Status:       constant.TransferStatusSuccess,
			},
			requesterMerchant: merchantRecipientID.String(),
			expected: &ListTransferResponse{
				UUID:          "00000000-0000-0000-0000-000000000000",
				ReferenceID:   "ref-id",
				Type:          constant.TransferTypeOUT,
				SenderID:      merchantSenderID.String(),
				SenderName:    "merchant 1",
				RecipientID:   merchantRecipientID.String(),
				RecipientName: "merchant 2",
				TransferType:  constant.MoneyFlowDirect,
				Amount:        100,
				Status:        constant.TransferStatusSuccess,
				Remarks:       "",
				UpdatedAt:     time.Time{},
				CreatedAt:     time.Time{},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := tc.input.ToListTransferResponse(tc.requesterMerchant, participantMap)
			assert.EqualValues(t, tc.expected, output)
		})
	}

}

func Test(t *testing.T) {
}
