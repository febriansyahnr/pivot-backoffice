package walletTransaction_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	walletTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/wallet/transaction"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/wallet/transaction"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantTransactionHistoryList(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db)

	transactionList := []walletTransactionModel.MerchantTransactionHistoryListResp{
		{
			Id:     "3b663b46-2bd1-47c8-a3c5-c00da1be9581",
			Type:   "MERCHANT_PAYMENT",
			Status: constant.StatusSuccess,
		},
	}
	tests := []struct {
		name          string
		request       walletTransactionModel.MerchantTransactionHistoryListReq
		setupMock     func()
		wantErr       error
		wantTotalRows int64
		wantResult    []walletTransactionModel.MerchantTransactionHistoryListResp
	}{
		{
			name: "ERROR:Count total rows",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Transaction history list",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
			wantResult: []walletTransactionModel.MerchantTransactionHistoryListResp{},
		},
		{
			name: "SUCCESS:Data not found with custom filter",
			request: walletTransactionModel.MerchantTransactionHistoryListReq{
				Id:     "3b663b46-2bd1-47c8-a3c5-c00da1be9581",
				Type:   "FEE_BILL_PAYMENT", // NOSONAR
				Status: constant.StatusSuccess,
			},
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
			wantResult: []walletTransactionModel.MerchantTransactionHistoryListResp{},
		},
		{
			name: "SUCCESS:Data found with custom filter",
			request: walletTransactionModel.MerchantTransactionHistoryListReq{
				ReferenceId: "REF",              // NOSONAR
				Type:        "MERCHANT_PAYMENT", // NOSONAR
				Status:      constant.StatusSuccess,
			},
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 1
				}).Return(nil)
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]walletTransactionModel.MerchantTransactionHistoryListResp) = transactionList
				}).Return(nil)
			},
			wantTotalRows: 1,
			wantResult:    transactionList,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, rows, err := repo.GetMerchantTransactionHistoryList(context.Background(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantTotalRows, rows)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetMerchantTransactionHistoryListForExport(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db)

	transactionList := []walletTransactionModel.MerchantTransactionHistoryListResp{
		{
			Id:     "3b663b46-2bd1-47c8-a3c5-c00da1be9581",
			Type:   "MERCHANT_PAYMENT",
			Status: constant.StatusSuccess,
		},
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []walletTransactionModel.MerchantTransactionHistoryListResp
	}{
		{
			name: "ERROR:Some Error",
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]walletTransactionModel.MerchantTransactionHistoryListResp) = transactionList
				}).Return(nil)
			},
			wantResult: transactionList,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetMerchantTransactionHistoryListForExport(context.Background(), walletTransactionModel.MerchantTransactionHistoryListReq{
				Status: "SUCCESS",
			})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetMerchantTransactionDetail(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db)

	merchantId := "441cbfbb-f58d-4bdd-9a65-31b273983ecc"
	transactionId := "896f09f4-8ac3-4e19-9c76-9b0bd80fcec8"
	result := walletTransactionModel.MerchantTransactionDetailResp{
		Id:                transactionId,
		RawAdditionalInfo: types.JSONText(`{}`),
		AdditionalInfo:    map[string]any{},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *walletTransactionModel.MerchantTransactionDetailResp
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, transactionId,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, transactionId,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, transactionId,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*walletTransactionModel.MerchantTransactionDetailResp) = result
				}).Return(nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetMerchantTransactionDetail(context.Background(), merchantId, transactionId)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
