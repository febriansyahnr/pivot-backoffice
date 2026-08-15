package unifiedPaymentModel

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/stretchr/testify/assert"
)

func TestDashboardPaymentLinkCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request DashboardPaymentLinkCreateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "success - valid request",
			request: DashboardPaymentLinkCreateRequest{
				Amount:    Amount{Value: 50000},
				ExpiredAt: time.Now().Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name: "error - amount below minimum",
			request: DashboardPaymentLinkCreateRequest{
				Amount:    Amount{Value: 5000},
				ExpiredAt: time.Now().Add(time.Hour),
			},
			wantErr: true,
			errMsg:  constant.ErrPaymentLinkMinAmount.Error(),
		},
		{
			name: "error - amount exceed maximum",
			request: DashboardPaymentLinkCreateRequest{
				Amount:    Amount{Value: constant.DashboardPaymentLinkMaxAmount + 1},
				ExpiredAt: time.Now().Add(time.Hour),
			},
			wantErr: true,
			errMsg:  constant.ErrPaymentLinkMaxAmount.Error(),
		},
		{
			name: "error - amount exactly at minimum boundary",
			request: DashboardPaymentLinkCreateRequest{
				Amount:    Amount{Value: constant.DashboardPaymentLinkMinAmount},
				ExpiredAt: time.Now().Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name: "error - expired time in past",
			request: DashboardPaymentLinkCreateRequest{
				Amount:    Amount{Value: 50000},
				ExpiredAt: time.Now().Add(-time.Hour),
			},
			wantErr: true,
			errMsg:  constant.ErrMsgExpiryLessThanCurrentTime,
		},
		{
			name: "error - both validations fail - should return first error",
			request: DashboardPaymentLinkCreateRequest{
				Amount:    Amount{Value: 5000},
				ExpiredAt: time.Now().Add(-time.Hour),
			},
			wantErr: true,
			errMsg:  constant.ErrPaymentLinkMinAmount.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnifiedPaymentSessionResponse_ToDashboardPaymentLinkResponse(t *testing.T) {
	now := time.Now()
	expiry := now.Add(time.Hour)

	tests := []struct {
		name     string
		input    UnifiedPaymentSessionResponse
		expected DashboardPaymentLinkResponse
	}{
		{
			name: "success - with expiry time",
			input: UnifiedPaymentSessionResponse{
				ID:                "session-123",
				ClientReferenceID: "ref-123",
				Amount:            Amount{Value: 50000},
				PaymentUrl:        "https://payment.link/123",
				ShortPaymentUrl:   "https://payment.link/123",
				Status:            "ACTIVE",
				CreatedAt:         now,
				ExpiryAt:          &expiry,
			},
			expected: DashboardPaymentLinkResponse{
				ID:                "session-123",
				ClientReferenceID: "ref-123",
				Amount:            Amount{Value: 50000},
				PaymentLink:       "https://payment.link/123",
				Status:            "ACTIVE",
				CreatedAt:         now,
				ExpiryAt:          expiry,
			},
		},
		{
			name: "success - without expiry time",
			input: UnifiedPaymentSessionResponse{
				ID:                "session-456",
				ClientReferenceID: "ref-456",
				Amount:            Amount{Value: 100000},
				PaymentUrl:        "https://payment.link/456",
				ShortPaymentUrl:   "https://payment.link/456",
				Status:            "PENDING",
				CreatedAt:         now,
				ExpiryAt:          nil,
			},
			expected: DashboardPaymentLinkResponse{
				ID:                "session-456",
				ClientReferenceID: "ref-456",
				Amount:            Amount{Value: 100000},
				PaymentLink:       "https://payment.link/456",
				Status:            "PENDING",
				CreatedAt:         now,
				ExpiryAt:          time.Time{}, // zero value when ExpiryAt is nil
			},
		},
		{
			name: "success - empty values",
			input: UnifiedPaymentSessionResponse{
				ID:                "",
				ClientReferenceID: "",
				Amount:            Amount{Value: 0},
				PaymentUrl:        "",
				Status:            "",
				CreatedAt:         time.Time{},
				ExpiryAt:          nil,
			},
			expected: DashboardPaymentLinkResponse{
				ID:                "",
				ClientReferenceID: "",
				Amount:            Amount{Value: 0},
				PaymentLink:       "",
				Status:            "",
				CreatedAt:         time.Time{},
				ExpiryAt:          time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToDashboardPaymentLinkResponse()

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.ClientReferenceID, result.ClientReferenceID)
			assert.Equal(t, tt.expected.Amount, result.Amount)
			assert.Equal(t, tt.expected.PaymentLink, result.PaymentLink)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.ExpiryAt, result.ExpiryAt)
		})
	}
}
