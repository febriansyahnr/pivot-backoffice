package qris_test

import (
	"database/sql"
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/qris"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationList(t *testing.T) {
	acquirerMerchantId := "8"
	source := Registration{
		Id:                       "1",
		ExternalId:               "2",
		Acquirer:                 "3",
		MerchantType:             "4",
		AcquirerParentMerchantId: "5",
		MerchantName:             "6",
		Status:                   "7",
		AcquirerMerchantId:       &acquirerMerchantId,
		CallbackDetailRaw: types.NullJSONText{
			Valid:    true,
			JSONText: []byte(`{"applymentCode":"1","resultCode":"2"}`),
		},
		CallbackDatetime: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(time.Minute),
	}

	want := RegistrationListResp{
		Id:                       "1",
		ExternalId:               "2",
		Acquirer:                 "3",
		MerchantType:             "4",
		AcquirerMerchantParentId: "5",
		MerchantName:             "6",
		Status:                   "7",
		AcquirerMerchantId:       source.AcquirerMerchantId,
		CallbackDetail: &CallbackDetail{
			ApplymentCode: "1", ResultCode: "2",
		},
		CallbackDatetime: &source.CallbackDatetime.Time,
		CreatedAt:        source.CreatedAt,
		UpdatedAt:        source.UpdatedAt,
	}
	actual := RegistrationListResp{}
	actual.FromRegistration(source)

	assert.Equal(t, want, actual)
}
