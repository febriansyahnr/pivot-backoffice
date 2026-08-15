package industry

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	t.Run("SUCCESS: Create new repository", func(t *testing.T) {
		mysqlMock := &mysqlMocks.IMySqlExt{}
		loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		repo := New(mysqlMock, loggerMock)

		assert.NotNil(t, repo)
	})
}

func TestGetAllIndustries(t *testing.T) {
	ctx := context.Background()

	sampleIndustries := []*industryModel.Industry{
		{
			UUID:           "uuid-1",
			ParentIndustry: "Technology",
			ChildIndustry:  "Software",
			RiskLevel:      "Low",
			MCC:            "5734",
			CommonMCC:      "5734",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			UUID:           "uuid-2",
			ParentIndustry: "Technology",
			ChildIndustry:  "Hardware",
			RiskLevel:      "Medium",
			MCC:            "5735",
			CommonMCC:      "5735",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	testCases := []struct {
		name      string
		request   *industryModel.SearchIndustryRequest
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  []*industryModel.Industry
		wantErr   bool
	}{
		{
			name:    "SUCCESS: Get all industries without keyword",
			request: &industryModel.SearchIndustryRequest{Keyword: ""},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					industries := args.Get(1).(*[]*industryModel.Industry)
					*industries = sampleIndustries
				})
			},
			expected: sampleIndustries,
			wantErr:  false,
		},
		{
			name:    "SUCCESS: Get industries with keyword",
			request: &industryModel.SearchIndustryRequest{Keyword: "Tech"},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					industries := args.Get(1).(*[]*industryModel.Industry)
					*industries = sampleIndustries[:1]
				})
			},
			expected: sampleIndustries[:1],
			wantErr:  false,
		},
		{
			name:    "SUCCESS: No results found",
			request: &industryModel.SearchIndustryRequest{Keyword: "NonExistent"},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name:    "ERROR: Database error",
			request: &industryModel.SearchIndustryRequest{Keyword: ""},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			result, err := repo.GetAllIndustries(ctx, tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestGetUniqueParentIndustries(t *testing.T) {
	ctx := context.Background()
	expectedParents := []string{"Technology", "Finance", "Healthcare"}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  []string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get unique parent industries",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					parents := args.Get(1).(*[]string)
					*parents = expectedParents
				})
			},
			expected: expectedParents,
			wantErr:  false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			result, err := repo.GetUniqueParentIndustries(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestGetChildIndustries(t *testing.T) {
	ctx := context.Background()
	parentIndustry := "Technology"
	expectedChildren := []string{"Software", "Hardware", "AI"}

	testCases := []struct {
		name           string
		parentIndustry string
		mockSetup      func(mysqlMock *mysqlMocks.IMySqlExt)
		expected       []string
		wantErr        bool
	}{
		{
			name:           "SUCCESS: Get child industries for parent",
			parentIndustry: parentIndustry,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					children := args.Get(1).(*[]string)
					*children = expectedChildren
				})
			},
			expected: expectedChildren,
			wantErr:  false,
		},
		{
			name:           "ERROR: Database error",
			parentIndustry: parentIndustry,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			result, err := repo.GetChildIndustries(ctx, tc.parentIndustry)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestGetMCCForIndustry(t *testing.T) {
	ctx := context.Background()
	parentIndustry := "Technology"
	childIndustry := "Software"
	expectedMCC := "5734"

	testCases := []struct {
		name           string
		parentIndustry string
		childIndustry  string
		mockSetup      func(mysqlMock *mysqlMocks.IMySqlExt)
		expected       string
		wantErr        bool
	}{
		{
			name:           "SUCCESS: Get MCC for industry combination",
			parentIndustry: parentIndustry,
			childIndustry:  childIndustry,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					mcc := args.Get(1).(*string)
					*mcc = expectedMCC
				})
			},
			expected: expectedMCC,
			wantErr:  false,
		},
		{
			name:           "ERROR: Database error",
			parentIndustry: parentIndustry,
			childIndustry:  childIndustry,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			result, err := repo.GetMCCForIndustry(ctx, tc.parentIndustry, tc.childIndustry)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestIsValidMCC(t *testing.T) {
	ctx := context.Background()
	mcc := "5734"

	testCases := []struct {
		name      string
		mcc       string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  bool
		wantErr   bool
	}{
		{
			name: "SUCCESS: Valid MCC found",
			mcc:  mcc,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					count := args.Get(1).(*int)
					*count = 1
				})
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Invalid MCC not found",
			mcc:  "9999",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					count := args.Get(1).(*int)
					*count = 0
				})
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "ERROR: Database error",
			mcc:  mcc,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(errors.New("database error"))
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			result, err := repo.IsValidMCC(ctx, tc.mcc)

			if tc.wantErr {
				assert.Error(t, err)
				assert.False(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestGetIndustryByID(t *testing.T) {
	ctx := context.Background()
	industryID := "uuid-1"

	sampleIndustry := &industryModel.Industry{
		UUID:           industryID,
		ParentIndustry: "Technology",
		ChildIndustry:  "Software",
		RiskLevel:      "Low",
		MCC:            "5734",
		CommonMCC:      "5734",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	testCases := []struct {
		name      string
		id        string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *industryModel.Industry
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get industry by ID",
			id:   industryID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					industry := args.Get(1).(*industryModel.Industry)
					*industry = *sampleIndustry
				})
			},
			expected: sampleIndustry,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Industry not found",
			id:   "non-existent-id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database error",
			id:   industryID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			result, err := repo.GetIndustryByID(ctx, tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}
