package advanceairepository

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProfileDetail(t *testing.T) {
	type Mocker struct {
		cfg         *config.Config
		secret      *config.Secret
		httpRequest *httpMocks.IHTTPRequest
	}

	mockSecret := &config.Secret{
		AdvanceAISecret: config.AdvanceAISecret{
			ApiKey: "test-api-key",
		},
	}

	mockConfig := &config.Config{
		AdvanceAIConfig: config.AdvanceAIConfig{
			BaseURL:   "https://api.advance.ai",
			JourneyID: "journey-123",
		},
	}

	validProfileResponse := &amlcommon.ProfileDetailResponse{
		Code:    "SUCCESS",
		Message: "SUCCESS",
		Data: amlcommon.ProfileDetailData{
			ProfileID: "e_tr_wci_1224148",
			ResultID:  "5jb7bg501ka71jw0o2z8oxkau",
			Name:      "Ir Joko WIDODO",
			AliasName: []string{"Mulyono", "Jokowi"},
			HitCategory: []string{"PEP"},
			DateOfBirth: "1961-06-21",
			Score: 100,
		},
		TransactionID: "4db50c53378fd2e0",
		Datetime:      1753868856.320797000,
		Timestamp:     1753868856320,
	}

	testCases := []struct {
		desc         string
		wantErr      bool
		mocker       func(m *Mocker)
		inquiryID    string
		profileID    string
		expcResponse *amlcommon.ProfileDetailResponse
	}{
		{
			desc:    "error request profile detail",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return(nil, 0, assert.AnError)
			},
			inquiryID:    "inquiry-123",
			profileID:    "profile-456",
			expcResponse: nil,
		},
		{
			desc:    "error validate response",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`INVALID JSON`), 200, nil)
			},
			inquiryID:    "inquiry-123",
			profileID:    "profile-456",
			expcResponse: nil,
		},
		{
			desc:    "HTTP error with valid response body",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "SUCCESS",
					"data": {
						"profileId": "e_tr_wci_1224148",
						"resultId": "5jb7bg501ka71jw0o2z8oxkau",
						"name": "Ir Joko WIDODO",
						"aliasName": ["Mulyono", "Jokowi"],
						"hitCategory": ["PEP"],
						"dateOfBirth": "1961-06-21",
						"score": 100,
						"profileRecordInfo": {
							"actions": [],
							"active": true,
							"addresses": [],
							"associates": [],
							"category": "POLITICAL INDIVIDUAL",
							"entityType": "INDIVIDUAL"
						}
					},
					"transactionId": "4db50c53378fd2e0",
					"datetime": 1753868856.320797000,
					"timestamp": 1753868856320
				}`), 400, assert.AnError)
			},
			inquiryID:    "inquiry-123",
			profileID:    "e_tr_wci_1224148",
			expcResponse: validProfileResponse,
		},
		{
			desc:    "HTTP error with invalid response body",
			wantErr: true,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`INVALID JSON`), 400, assert.AnError)
			},
			inquiryID:    "inquiry-123",
			profileID:    "profile-456",
			expcResponse: nil,
		},
		{
			desc:    "success when response is 200",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "SUCCESS",
					"data": {
						"profileId": "e_tr_wci_1224148",
						"resultId": "5jb7bg501ka71jw0o2z8oxkau",
						"name": "Ir Joko WIDODO",
						"aliasName": ["Mulyono", "Jokowi"],
						"hitCategory": ["PEP"],
						"dateOfBirth": "1961-06-21",
						"score": 100,
						"profileRecordInfo": {
							"actions": [],
							"active": true,
							"addresses": [],
							"associates": [],
							"category": "POLITICAL INDIVIDUAL",
							"entityType": "INDIVIDUAL"
						}
					},
					"transactionId": "4db50c53378fd2e0",
					"datetime": 1753868856.320797000,
					"timestamp": 1753868856320
				}`), 200, nil)
			},
			inquiryID:    "inquiry-123",
			profileID:    "e_tr_wci_1224148",
			expcResponse: validProfileResponse,
		},
		{
			desc:    "success with complex profile data",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "SUCCESS",
					"data": {
						"profileId": "e_tr_wci_9999999",
						"resultId": "test-result-id",
						"name": "Test Person",
						"aliasName": ["Alias1", "Alias2"],
						"hitCategory": ["PEP", "SAN"],
						"dateOfBirth": "1990-01-01",
						"score": 85,
						"profileRecordInfo": {
							"actions": [],
							"active": true,
							"addresses": [
								{
									"city": "Jakarta",
									"country": {
										"code": "IDN",
										"name": "INDONESIA"
									},
									"region": "Jakarta"
								}
							],
							"associates": [],
							"category": "INDIVIDUAL",
							"entityType": "INDIVIDUAL"
						}
					},
					"transactionId": "test-transaction-id",
					"datetime": 1234567890.123,
					"timestamp": 1234567890123
				}`), 200, nil)
			},
			inquiryID: "inquiry-test",
			profileID: "e_tr_wci_9999999",
			expcResponse: &amlcommon.ProfileDetailResponse{
				Code:    "SUCCESS",
				Message: "SUCCESS",
				Data: amlcommon.ProfileDetailData{
					ProfileID:   "e_tr_wci_9999999",
					ResultID:    "test-result-id",
					Name:        "Test Person",
					AliasName:   []string{"Alias1", "Alias2"},
					HitCategory: []string{"PEP", "SAN"},
					DateOfBirth: "1990-01-01",
					Score:       85,
					ProfileRecordInfo: amlcommon.ProfileRecordInfo{
						Actions:    []any{},
						Active:     true,
						Addresses:  []amlcommon.Address{
							{
								City: "Jakarta",
								Country: amlcommon.Country{
									Code: "IDN",
									Name: "INDONESIA",
								},
								Region: "Jakarta",
							},
						},
						Associates: []amlcommon.Associate{},
						Category:    "INDIVIDUAL",
						EntityType:  "INDIVIDUAL",
					},
				},
				TransactionID: "test-transaction-id",
				Datetime:      1234567890.123,
				Timestamp:     1234567890123,
			},
		},
		{
			desc:    "empty inquiry ID",
			wantErr: false,
			mocker: func(m *Mocker) {
				m.cfg = mockConfig
				m.secret = mockSecret
				m.httpRequest.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]string"),
				).Return([]byte(`{
					"code": "SUCCESS",
					"message": "SUCCESS",
					"data": {
						"profileId": "test-profile",
						"name": "Test",
						"profileRecordInfo": {
							"entityType": "INDIVIDUAL"
						}
					}
				}`), 200, nil)
			},
			inquiryID: "",
			profileID: "test-profile",
			expcResponse: &amlcommon.ProfileDetailResponse{
				Code:    "SUCCESS",
				Message: "SUCCESS",
				Data: amlcommon.ProfileDetailData{
					ProfileID: "test-profile",
					Name:      "Test",
					ProfileRecordInfo: amlcommon.ProfileRecordInfo{
						EntityType: "INDIVIDUAL",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			m := &Mocker{
				cfg:    &config.Config{},
				secret: &config.Secret{},

				httpRequest: httpMocks.NewIHTTPRequest(t),
			}

			tc.mocker(m)

			r := New(m.cfg, m.secret, log, m.httpRequest)

			resp, err := r.ProfileDetail(context.Background(), tc.inquiryID, tc.profileID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expcResponse != nil {
				assert.NotNil(t, resp)
				assert.Equal(t, tc.expcResponse.Code, resp.Code)
				assert.Equal(t, tc.expcResponse.Message, resp.Message)
				assert.Equal(t, tc.expcResponse.Data.ProfileID, resp.Data.ProfileID)
				assert.Equal(t, tc.expcResponse.Data.Name, resp.Data.Name)
				assert.Equal(t, tc.expcResponse.TransactionID, resp.TransactionID)
			}

			m.httpRequest.AssertExpectations(t)
		})
	}
}