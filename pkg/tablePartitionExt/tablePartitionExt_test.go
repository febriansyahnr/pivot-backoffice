package tablePartitionExt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name    string
		input   PartitionConfig
		wantErr error
	}{
		{
			name: "When table name is empty, should return table name is required error",
			input: PartitionConfig{
				TableName:      "",
				TotalPartition: 1,
				StartedAt:      time.Now(),
			},
			wantErr: errors.New("table name is required"),
		},
		{
			name: "When total partition is zero, should return total partition is required error",
			input: PartitionConfig{
				TableName:      "test_table",
				TotalPartition: 0,
				StartedAt:      time.Now(),
			},
			wantErr: errors.New("total partition is required"),
		},
		{
			name: "When startedAt is zero, should return start time is required error",
			input: PartitionConfig{
				TableName:      "test_table",
				TotalPartition: 1,
				StartedAt:      time.Time{},
			},
			wantErr: errors.New("start time is required"),
		},
		{
			name: "When all fields are valid, should return no error",
			input: PartitionConfig{
				TableName:      "test_table",
				TotalPartition: 3,
				StartedAt:      time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pt := partitionTable{}
			err := pt.validateConfig(tc.input)

			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestGetTotalPartitions(t *testing.T) {
	testCases := []struct {
		name      string
		cfg       PartitionConfig
		mockSetup func(db *mysqlMock.IMySqlExt)
		wantError error
		want      int
	}{
		{
			name: "When query is successful, should return total partitions",
			cfg:  PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 3
				})
			},
			wantError: nil,
			want:      3,
		},
		{
			name: "When query fails, should return error",
			cfg:  PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(errors.New("database error"))
			},
			wantError: errors.New("database error"),
			want:      0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMock.NewIMySqlExt(t)
			tc.mockSetup(mockDB)

			pt := partitionTable{
				db:          mockDB,
				tableSchema: "test_schema",
			}

			ctx := context.Background()
			result, err := pt.getTotalPartitions(ctx, tc.cfg)

			assert.Equal(t, tc.want, result)
			assert.Equal(t, tc.wantError, err)

			mockDB.AssertExpectations(t)
		})
	}
}

func TestIsPartitionExist(t *testing.T) {
	testCases := []struct {
		name      string
		payload   string
		cfg       PartitionConfig
		mockSetup func(db *mysqlMock.IMySqlExt)
		wantError error
		want      bool
	}{
		{
			name:    "When partition exist, should return true",
			payload: "partition-1",
			cfg:     PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(string(constant.PtrBoolMockType())),
					constant.StringMockType(),
					"test_schema",
					"test_table",
					"partition-1",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*bool) = true
				})
			},
			wantError: nil,
			want:      true,
		},
		{
			name:    "When partition not exist, should return false",
			payload: "partition-2",
			cfg:     PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(string(constant.PtrBoolMockType())),
					constant.StringMockType(),
					"test_schema",
					"test_table",
					"partition-2",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*bool) = false
				})
			},
			wantError: nil,
			want:      false,
		},
		{
			name:    "When query fails, should return error",
			payload: "invalid-partition",
			cfg:     PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(string(constant.PtrBoolMockType())),
					constant.StringMockType(),
					"test_schema",
					"test_table",
					"invalid-partition",
				).Return(errors.New("database error"))
			},
			wantError: errors.New("database error"),
			want:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMock.NewIMySqlExt(t)
			tc.mockSetup(mockDB)

			pt := partitionTable{
				db:          mockDB,
				tableSchema: "test_schema",
			}

			ctx := context.Background()
			result, err := pt.isPartitionExist(ctx, tc.cfg, tc.payload)

			assert.Equal(t, tc.want, result)
			assert.Equal(t, tc.wantError, err)

			mockDB.AssertExpectations(t)
		})
	}
}

func TestInitDayRangePartition(t *testing.T) {
	loggerMock := loggerMock.NewILogger(t)
	loggerMock.On("Error", mock.Anything, "error initializing table partitions", mock.Anything)

	testCases := []struct {
		name      string
		cfg       PartitionConfig
		mockSetup func(db *mysqlMock.IMySqlExt)
		wantError error
	}{
		{
			name: "When success to create partition, then should not return error",
			cfg:  PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "When partition not exist, should return false",
			cfg:  PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("duplicate partition"))
			},
			wantError: errors.New("duplicate partition"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMock.NewIMySqlExt(t)
			tc.mockSetup(mockDB)

			pt := partitionTable{
				db:          mockDB,
				log:         loggerMock,
				tableSchema: "test_schema",
			}

			ctx := context.Background()
			err := pt.initDayRangePartition(ctx, tc.cfg)

			assert.Equal(t, tc.wantError, err)

			mockDB.AssertExpectations(t)
		})
	}
}

func TestManageDefaultMaxPartition(t *testing.T) {
	pt := partitionTable{}
	res := pt.ManageDefaultMaxPartition(context.Background(), PartitionConfig{})
	assert.Equal(t, nil, res)
}

func TestCreateDayPartitions(t *testing.T) {
	var (
		ctx = context.Background()
	)
	testCases := []struct {
		name      string
		payload   PartitionConfig
		setupMock func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger)
		wantErr   error
	}{
		{
			name: "when the partition was 0, then should return nil",
			payload: PartitionConfig{
				TotalPartition: 0,
				TableName:      "test_table",
				StartedAt:      time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC),
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				log.On("Info", mock.Anything, "no new partitions to add, skipping")
			},
		},
		{
			name: "when failed to check existing partition, then should return error",
			payload: PartitionConfig{
				TableName:      "sample_table",
				TotalPartition: 1,
				StartedAt:      time.Now().UTC(),
				Parameter:      "created_at",
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
				).Return(sql.ErrNoRows)

				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(string(constant.PtrBoolMockType())),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
					mock.AnythingOfType(string(constant.StringMockType())),
				).Return(errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "when the partition exist, then should not return error",
			payload: PartitionConfig{
				TableName:      "sample_table",
				TotalPartition: 1,
				StartedAt:      time.Now().UTC(),
				Parameter:      "created_at",
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
				).Return(sql.ErrNoRows)

				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(string(constant.PtrBoolMockType())),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
					mock.AnythingOfType(string(constant.StringMockType())),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*bool) = true
				})
				log.On("Info", mock.Anything, "no new partitions to add, skipping")
			},
		},
		{
			name: "when failed create new partition, then should return error",
			payload: PartitionConfig{
				TableName:      "sample_table",
				TotalPartition: 1,
				StartedAt:      time.Now().UTC(),
				Parameter:      "created_at",
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
				).Return(sql.ErrNoRows)

				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(string(constant.PtrBoolMockType())),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
					mock.AnythingOfType(string(constant.StringMockType())),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*bool) = false
				})

				db.On(
					"ExecContext",
					mock.Anything,
					constant.StringMockType(),
				).Return(false, errors.New("database error"))

				log.On("Error", mock.Anything, "error when adding new partitions", mock.Anything)
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "when success create new partition, then should not return error",
			payload: PartitionConfig{
				TableName:      "sample_table",
				TotalPartition: 1,
				StartedAt:      time.Now().UTC(),
				Parameter:      "created_at",
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
				).Return(sql.ErrNoRows)

				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(string(constant.PtrBoolMockType())),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
					mock.AnythingOfType(string(constant.StringMockType())),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*bool) = false
				})

				db.On(
					"ExecContext",
					mock.Anything,
					constant.StringMockType(),
				).Return(true, nil)
			},
		},
		{
			name: "when failed to get latest database partition, then should return error",
			payload: PartitionConfig{
				TotalPartition: 1,
				TableName:      "test_table",
				StartedAt:      time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC),
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"backend_portal_test",
					"test_table",
				).Return(fmt.Errorf("database error"))
				log.On("Error", mock.Anything, "err-get-latest-partition", mock.Anything, mock.Anything)
			},
			wantErr: fmt.Errorf("database error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMock.NewIMySqlExt(t)
			loggerMock := loggerMock.NewILogger(t)

			tc.setupMock(mockDB, loggerMock)

			pt := partitionTable{
				db:          mockDB,
				log:         loggerMock,
				tableSchema: "backend_portal_test",
			}

			err := pt.createDayPartitions(ctx, tc.payload)

			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestCreateDayRangePartition(t *testing.T) {
	var (
		ctx = context.Background()
	)
	testCases := []struct {
		name      string
		payload   PartitionConfig
		setupMock func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger)
		wantErr   error
	}{
		{
			name:    "when config was invalid then should return error",
			payload: PartitionConfig{},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
			},
			wantErr: errors.New("table name is required"),
		},
		{
			name: "when failed to get existing partition, then should return error",
			payload: PartitionConfig{
				TableName:      "sample_table",
				TotalPartition: 1,
				StartedAt:      time.Now().UTC(),
				Parameter:      "created_at",
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
				).Return(errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "when failed init partition, then should return error",
			payload: PartitionConfig{
				TableName:      "sample_table",
				TotalPartition: 1,
				StartedAt:      time.Now().UTC(),
				Parameter:      "created_at",
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 0
				})

				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("duplicate partition"))

				log.On("Error", mock.Anything, "error initializing table partitions", mock.Anything)

			},
			wantErr: errors.New("duplicate partition"),
		},
		{
			name: "when success init partition, then should not error",
			payload: PartitionConfig{
				TableName:      "sample_table",
				TotalPartition: 1,
				StartedAt:      time.Now().UTC(),
				Parameter:      "created_at",
			},
			setupMock: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					"backend_portal_test",
					"sample_table",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 0
				})

				db.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
				).Return(true, nil)

				log.On("Info", mock.Anything, "no new partitions to add, skipping")

			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMock.NewIMySqlExt(t)
			loggerMock := loggerMock.NewILogger(t)

			tc.setupMock(mockDB, loggerMock)

			pt := New(mockDB, loggerMock, "backend_portal_test")
			err := pt.CreateDayRangePartition(ctx, tc.payload)

			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestGetLatestPartitionName(t *testing.T) {
	testCases := []struct {
		name      string
		cfg       PartitionConfig
		mockSetup func(db *mysqlMock.IMySqlExt)
		wantError error
		want      string
	}{
		{
			name: "When latest partition found, should return the name",
			cfg:  PartitionConfig{TableName: "test_table", StartedAt: time.Now()},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*string) = "p20210101"
				})
			},
			wantError: nil,
			want:      "p20210101",
		},
		{
			name: "When partition not exist, should return empty slice",
			cfg:  PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(sql.ErrNoRows)
			},
			wantError: nil,
			want:      "",
		},
		{
			name: "When query fails, should return error",
			cfg:  PartitionConfig{TableName: "test_table"},
			mockSetup: func(db *mysqlMock.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(errors.New("database error"))
			},
			wantError: errors.New("database error"),
			want:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMock.NewIMySqlExt(t)
			tc.mockSetup(mockDB)

			pt := partitionTable{
				db:          mockDB,
				tableSchema: "test_schema",
			}

			ctx := context.Background()
			result, err := pt.getLatestPartitionName(ctx, tc.cfg)

			assert.Equal(t, tc.want, result)
			assert.Equal(t, tc.wantError, err)

			mockDB.AssertExpectations(t)
		})
	}
}
func TestCreateMissingPartitionRangeDefinition(t *testing.T) {
	var (
		ctx = context.Background()
	)

	testCases := []struct {
		name      string
		cfg       PartitionConfig
		mockSetup func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger)
		want      []string
		wantErr   error
	}{
		{
			name: "When failed to get latest partition, should return error",
			cfg: PartitionConfig{
				TableName: "test_table",
				StartedAt: time.Now().UTC(),
			},
			mockSetup: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(errors.New("database error"))
				log.On("Error", mock.Anything, "err-get-latest-partition", mock.Anything, mock.Anything)
			},
			want:    nil,
			wantErr: errors.New("database error"),
		},
		{
			name: "When no latest partition, should return empty slice",
			cfg: PartitionConfig{
				TableName: "test_table",
				StartedAt: time.Now().UTC(),
			},
			mockSetup: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(sql.ErrNoRows)
			},
			want:    []string{},
			wantErr: nil,
		},
		{
			name: "When there are missing partitions, should return the missing partitions",
			cfg: PartitionConfig{
				TableName: "test_table",
				StartedAt: time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC),
			},
			mockSetup: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*string) = "p20231008"
				})
			},
			want: []string{
				"PARTITION p20231009 VALUES LESS THAN (UNIX_TIMESTAMP('2023-10-10'))",
			},
			wantErr: nil,
		},
		{
			name: "When no missing partitions, should return empty slice",
			cfg: PartitionConfig{
				TableName: "test_table",
				StartedAt: time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC),
			},
			mockSetup: func(db *mysqlMock.IMySqlExt, log *loggerMock.ILogger) {
				db.On(
					"GetContext",
					mock.Anything,
					constant.PtrStringMockType(),
					constant.StringMockType(),
					"test_schema",
					"test_table",
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*string) = "p20231010"
				})
			},
			want:    []string{},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMock.NewIMySqlExt(t)
			loggerMock := loggerMock.NewILogger(t)

			tc.mockSetup(mockDB, loggerMock)

			pt := partitionTable{
				db:          mockDB,
				log:         loggerMock,
				tableSchema: "test_schema",
			}

			result, err := pt.createMissingPartitionRangeDefinition(ctx, tc.cfg)

			assert.Equal(t, tc.want, result)
			assert.Equal(t, tc.wantErr, err)

			mockDB.AssertExpectations(t)
			loggerMock.AssertExpectations(t)
		})
	}
}
