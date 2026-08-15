package location_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/location"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/location"
	mySqlExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetProvinces(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)
	provinces := []location.Province{
		{
			Id:        15,
			Name:      "JAWA TIMUR",
			CreatedAt: time.Now().UTC(),
		},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult []location.Province
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]location.Province"), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]location.Province"), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]location.Province)) = provinces
				}).Return(nil)
			},
			wantResult: provinces,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if resp, err := repo.GetProvinces(context.Background()); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetCitiesByProvinceId(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)
	cities := []location.City{
		{
			Id:         259,
			ProvinceId: 15,
			Name:       "KOTA MALANG",
			CreatedAt:  time.Now().UTC(),
		},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult []location.City
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]location.City"), c.StringMockType(), c.Uint16MockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]location.City"), c.StringMockType(), c.Uint16MockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]location.City)) = cities
				}).Return(nil)
			},
			wantResult: cities,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if resp, err := repo.GetCitiesByProvinceId(context.Background(), 15); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetDistrictsByCityId(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)
	districts := []location.District{
		{
			Id:        3887,
			CityId:    259,
			Name:      "BLIMBING",
			CreatedAt: time.Now().UTC(),
		},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult []location.District
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]location.District"), c.StringMockType(), c.Uint16MockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]location.District"), c.StringMockType(), c.Uint16MockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]location.District)) = districts
				}).Return(nil)
			},
			wantResult: districts,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if resp, err := repo.GetDistrictsByCityId(context.Background(), 259); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetDistrictById(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)
	district := &location.District{
		Id:     254,
		CityId: 1,
		Name:   "SUKUN",
	}
	ptrDistrictMockType := mock.AnythingOfType("*location.District")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult *location.District
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrDistrictMockType, c.StringMockType(), c.Uint16MockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrDistrictMockType, c.StringMockType(), c.Uint16MockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrDistrictMockType, c.StringMockType(), c.Uint16MockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*location.District)) = *district
				}).Return(nil)
			},
			wantResult: district,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if resp, err := repo.GetDistrictById(context.Background(), 259); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
