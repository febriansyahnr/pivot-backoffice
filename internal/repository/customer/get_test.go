package customerRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCustomerList(t *testing.T) {
	testCases := []struct {
		desc        string
		wantErr     bool
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		phoneNumber string
	}{
		{
			desc: "SUCCESS: get customer list",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			desc: "SUCCESS: get customer list with phone number filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			phoneNumber: "081234567890",
		},
		{
			desc: "SUCCESS: get customer list with results that need transformation",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					// Set total count
					totalPtr := args.Get(1).(*int64)
					*totalPtr = 2
				}).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					// Get the customers slice and populate it
					customers := args.Get(1).(*[]customerModel.CustomerDBModel)
					*customers = []customerModel.CustomerDBModel{
						{
							UUID:        "cust-123",
							MerchantID:  "merchant-id",
							PhoneNumber: "08123456789",
							FirstName:   "John",
							LastName: sql.NullString{
								String: "Doe",
								Valid:  true,
							},
							Email: sql.NullString{
								String: "john.doe@example.com",
								Valid:  true,
							},
							Metadata: []byte(`{"key":"value"}`),
						},
						{
							UUID:        "cust-456",
							MerchantID:  "merchant-id",
							PhoneNumber: "08987654321",
							FirstName:   "Jane",
							LastName: sql.NullString{
								String: "Smith",
								Valid:  true,
							},
							Metadata: []byte(`{"another_key":"another_value"}`),
						},
					}
				}).Return(nil)
			},
		},
		{
			desc: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			customers, meta, err := repo.GetCustomerList(ctx, "merchant-id", tc.phoneNumber, 1, 10)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Additional assertions for the transformation test case
				if tc.desc == "SUCCESS: get customer list with results that need transformation" {
					// Check number of results
					assert.Equal(t, 2, len(customers))

					// Check first customer was transformed correctly
					assert.Equal(t, "cust-123", customers[0].UUID)
					assert.Equal(t, "John", customers[0].FirstName)
					assert.Equal(t, "Doe", customers[0].LastName)
					assert.Equal(t, "john.doe@example.com", customers[0].Email)
					assert.Equal(t, map[string]interface{}{"key": "value"}, customers[0].Metadata)

					// Check second customer was transformed correctly
					assert.Equal(t, "cust-456", customers[1].UUID)
					assert.Equal(t, "Jane", customers[1].FirstName)
					assert.Equal(t, "Smith", customers[1].LastName)
					assert.Equal(t, map[string]interface{}{"another_key": "another_value"}, customers[1].Metadata)

					// Check meta information
					assert.Equal(t, int64(2), meta.TotalItems)
				}
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetById(t *testing.T) {
	email := sql.NullString{
		String: "VJ2jK@example.com",
		Valid:  true,
	}
	phone := "081234567890"

	validExpected := customerModel.CustomerDBModel{
		UUID:        "123",
		FirstName:   "John Doe",
		Email:       email,
		PhoneNumber: phone,
	}

	testCases := []struct {
		desc       string
		customerID string
		wantErr    bool
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		expected   *customerModel.CustomerDBModel
	}{
		{
			desc:       "SUCCESS: get customer by id",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					customer := args.Get(1).(*customerModel.CustomerDBModel)
					*customer = validExpected
				})
			},
			expected: &validExpected,
		},
		{
			desc:       "ERROR: Customer Not Found",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			desc:       "ERROR: Database Error",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			_, err := repo.GetCustomerById(ctx, tc.customerID, "merchantID")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetByPhoneNumber(t *testing.T) {
	phone := "081234567890"

	validExpected := customerModel.CustomerDBModel{
		UUID:        "123",
		FirstName:   "John Doe",
		PhoneNumber: phone,
	}

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *customerModel.CustomerDBModel
	}{
		{
			desc: "SUCCESS: get customer by email",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					customer := args.Get(1).(*customerModel.CustomerDBModel)
					*customer = validExpected
				})
			},
			expected: &validExpected,
		},
		{
			desc: "ERROR: Customer Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			desc: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			_, err := repo.GetCustomerByPhoneNumber(ctx, phone, "")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestFindByEmail(t *testing.T) {
	email := "VJ2jK@example.com"
	phone := "081234567890"

	validExpected := customerModel.CustomerDBModel{
		UUID:      "123",
		FirstName: "John Doe",
		Email: sql.NullString{
			String: email,
			Valid:  true,
		},
		PhoneNumber: phone,
	}

	testCases := []struct {
		desc          string
		customerEmail string
		wantErr       bool
		mockSetup     func(mysqlMock *mysqlMocks.IMySqlExt)
		expected      *customerModel.CustomerDBModel
	}{
		{
			desc:          "SUCCESS: get customer by email",
			customerEmail: email,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					customer := args.Get(1).(*customerModel.CustomerDBModel)
					*customer = validExpected
				})
			},
			expected: &validExpected,
		},
		{
			desc:          "ERROR: Customer Not Found",
			customerEmail: email,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			desc:          "ERROR: Database Error",
			customerEmail: email,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			_, err := repo.FindCustomerByEmail(ctx, tc.customerEmail)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestFindById(t *testing.T) {
	email := "VJ2jK@example.com"
	phone := "081234567890"

	validExpected := customerModel.CustomerDBModel{
		UUID:      "123",
		FirstName: "John Doe",
		Email: sql.NullString{
			String: email,
			Valid:  true,
		},
		PhoneNumber: phone,
	}

	testCases := []struct {
		desc       string
		customerID string
		wantErr    bool
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		expected   *customerModel.CustomerDBModel
	}{
		{
			desc:       "SUCCESS: get customer by id",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					customer := args.Get(1).(*customerModel.CustomerDBModel)
					*customer = validExpected
				})
			},
			expected: &validExpected,
		},
		{
			desc:       "ERROR: Customer Not Found",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			desc:       "ERROR: Database Error",
			customerID: "123",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			_, err := repo.FindCustomerById(ctx, tc.customerID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetMerchantCustomerByEmail(t *testing.T) {
	email := "VJ2jK@example.com"
	phone := "081234567890"
	merchantID := "merchant-id"

	validExpected := customerModel.CustomerDBModel{
		UUID:       "123",
		MerchantID: merchantID,
		FirstName:  "John Doe",
		Email: sql.NullString{
			String: email,
			Valid:  true,
		},
		PhoneNumber: phone,
	}

	testCases := []struct {
		desc      string
		request   customerModel.GetMerchantCustomerRequest
		wantErr   bool
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *customerModel.CustomerDBModel
	}{
		{
			desc:    "SUCCESS: get customer by email",
			request: customerModel.GetMerchantCustomerRequest{MerchantID: merchantID, Email: email},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.Anything,
					merchantID,
					email,
				).Return(nil).Run(func(args mock.Arguments) {
					customer := args.Get(1).(*customerModel.CustomerDBModel)
					*customer = validExpected
				}).Once()
			},
			expected: &validExpected,
		},
		{
			desc:    "ERROR: Customer Not Found",
			request: customerModel.GetMerchantCustomerRequest{MerchantID: merchantID, Email: "invalid-email"},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*customerModel.CustomerDBModel"),
					mock.Anything,
					merchantID,
					"invalid-email",
				).Return(sql.ErrNoRows).Once()
			},
			expected: nil,
			wantErr:  false,
		},
		{
			desc:    "ERROR: Database Error",
			request: customerModel.GetMerchantCustomerRequest{MerchantID: merchantID, Email: email},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					merchantID,
					email,
				).Return(errors.New("database error")).Once()
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			_, err := repo.GetMerchantCustomerByEmail(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetCardFundedPayoutSavedCardList(t *testing.T) {
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	testCases := []struct {
		desc      string
		filter    *cardFundedPayoutModel.FilterGetSavedCardList
		wantErr   bool
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
	}{
		{
			desc: "SUCCESS: get saved card list with default filter",
			filter: &cardFundedPayoutModel.FilterGetSavedCardList{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    1000,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			desc: "SUCCESS: get saved card list with results",
			filter: &cardFundedPayoutModel.FilterGetSavedCardList{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					totalPtr := args.Get(1).(*int64)
					*totalPtr = 1
				}).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					customers := args.Get(1).(*[]customerModel.CustomerDBModel)
					*customers = []customerModel.CustomerDBModel{
						{
							UUID:       "customer-123",
							MerchantID: "merchant-123",
							FirstName:  "John Doe",
							Metadata: []byte(`{
								"useCase":"CARD_FUNDED_PAYOUT_SAVED_CARDS",
								"paymentMethods":[{
									"paymentChannel":"CHANNEL",
									"card":{
										"cardName":"VISA",
										"network":"BANK",
										"last4":"1234",
										"expMonth":"12",
										"expYear":"2025"
									}
								}]
							}`),
						},
					}
				}).Return(nil)
			},
		},
		{
			desc: "SUCCESS: get saved card list with date filter",
			filter: &cardFundedPayoutModel.FilterGetSavedCardList{
				MerchantID:     "merchant-123",
				Page:           1,
				PerPage:        1000,
				Sort:           "ASC",
				SortBy:         "updatedAt",
				StartCreatedAt: &startDate,
				EndCreatedAt:   &endDate,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			desc: "ERROR: Database Error on GetContext",
			filter: &cardFundedPayoutModel.FilterGetSavedCardList{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    1000,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: true,
		},
		{
			desc: "ERROR: Database Error on SelectContext",
			filter: &cardFundedPayoutModel.FilterGetSavedCardList{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    1000,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "customers")
			result, err := repo.GetCardFundedPayoutSavedCardList(ctx, tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Data)
				assert.NotNil(t, result.Meta)

				// Additional assertions for results test case
				if tc.desc == "SUCCESS: get saved card list with results" {
					assert.Equal(t, int64(1), result.Meta.TotalItems)
					savedCards, ok := result.Data.([]cardFundedPayoutModel.GetSavedCardResponse)
					assert.True(t, ok)
					assert.Equal(t, 1, len(savedCards))
				}
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetCardFundedPayoutSavedCardDetail(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	var (
		cardID     = "52a3ba67-a79d-46fd-a681-99769329447f"
		merchantID = "0b5b90c5-17f2-485b-a3ec-727adc185680"
	)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *cardFundedPayoutModel.GetSavedCardResponse
	}{
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, cardID,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, cardID,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, cardID,
				).Once().Return(nil)
			},
			wantResult: &cardFundedPayoutModel.GetSavedCardResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetCardFundedPayoutSavedCardDetail(t.Context(), cardFundedPayoutModel.GetSavedCardDetailRequest{
				CardID:     cardID,
				MerchantID: merchantID,
			})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}
