package crmXbController

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func getSamplePayoutResponse(uuid string, merchantId string, created time.Time) *xbModel.GetPayoutResponse {
	return &xbModel.GetPayoutResponse{
		Uuid:                uuid,
		MerchantId:          merchantId,
		ReferenceId:         "ref_987654",
		SourceCurrency:      "USD",
		DestinationCurrency: "EUR",
		DestinationAmount:   decimal.NewFromFloat(850.75),
		FxRate:              decimal.NewFromFloat(1.12),
		DestinationFxRate:   decimal.NewFromFloat(1.14),
		Fee:                 decimal.NewFromFloat(15.00),
		TotalAmount:         decimal.NewFromFloat(865.75),
		Remark:              "Payment for invoice #12345",
		CreatedAt:           created,
		BeneficiaryData: xbModel.BeneficiaryDataResponse{
			Name:               "John Doe",
			CountryCode:        "US",
			CountryName:        "United States",
			State:              "California",
			City:               "Los Angeles",
			Address:            "1234 Sunset Blvd",
			Postcode:           "90001",
			AccountNumber:      "1234567890",
			BankName:           "Bank of America",
			BankCode:           "BOFAUS3N",
			ContactCountryCode: "+1",
			ContactNumber:      "5551234567",
			Email:              "johndoe@example.com",
		},
		SenderData: xbModel.SenderDataResponse{
			Name:                 "Jane Smith",
			CountryCode:          "GB",
			CountryName:          "United Kingdom",
			State:                "England",
			City:                 "London",
			Address:              "12 Oxford Street",
			Postcode:             "W1D 1AB",
			AccountType:          "Personal",
			IdentificationType:   "Passport",
			IdentificationNumber: "123456789",
			BankAccountNumber:    "9876543210",
			Dob:                  "1990-05-12",
			ContactCountryCode:   "+44",
			ContactNumber:        "7911123456",
			SourceOfIncome:       "Salary",
		},
		Status:            "PAID",
		StatusDescription: "Payout has been successfully completed",
		RfiDetails: []*xbModel.RfiDetails{
			{
				PartnerDocumentID:       "doc_001",
				PartnerDocumentEntityID: "entity_001",
				Actor:                   "BENEFICIARY",
				Entity:                  "bank_document",
				DocumentType:            "ID Proof",
				DocumentURL:             "https://example.com/docs/id_proof.pdf",
				Filename:                "id_proof.pdf",
				Value:                   "123456789",
				Comment:                 "Please review the ID proof",
				Status:                  "received",
				RequestedAt:             nil,
			},
			{
				PartnerDocumentID:       "doc_002",
				PartnerDocumentEntityID: "entity_002",
				Actor:                   "REMITTER",
				Entity:                  "address_document",
				DocumentType:            "Address Proof",
				DocumentURL:             "https://example.com/docs/address_proof.pdf",
				Filename:                "address_proof.pdf",
				Value:                   "987654321",
				Comment:                 "Address confirmation required",
				Status:                  "pending",
				RequestedAt:             nil,
			},
		},
	}
}

func TestGetPayouByID(t *testing.T) {
	svc := serviceMocks.NewIXbPayoutService(t)
	validPayoutID := uuid.NewString()
	validMerchantID := uuid.NewString()
	now := time.Now()

	validResponseInJson, err := json.Marshal(getSamplePayoutResponse(validPayoutID, validMerchantID, now))
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	tests := []struct {
		name           string
		payoutID       string
		setupBody      func(*testing.T) []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:     "ERROR: Invalid PayoutID format",
			payoutID: "invalid",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:     "ERROR: GetPayoutById service error",
			payoutID: validPayoutID,
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.GetPayoutRequest{
					PayoutId: validPayoutID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"GetPayoutById",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetPayoutRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:     "SUCCESS",
			payoutID: validPayoutID,
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.GetPayoutRequest{
					PayoutId:   validPayoutID,
					MerchantId: validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				svc.On(
					"GetPayoutById",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetPayoutRequest"),
				).Return(getSamplePayoutResponse(validPayoutID, validMerchantID, now), nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/xb/payout/%s", test.payoutID), bytes.NewBuffer(test.setupBody(t)))

			router := chi.NewRouter()
			router.Post("/xb/payout/{id}", New(svc).GetPayoutByID)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
