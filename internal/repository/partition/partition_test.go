package partitionRepository_test

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/partition"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"
	"github.com/paper-indonesia/pdk/v2/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var tzAsiaJakarta, _ = time.LoadLocation(constant.TimeLoc)

func TestReorganizeDailyRangePartition(t *testing.T) {
	mysql := mocks.NewIMySqlExt(t)

	db := New(mysql)

	requestDatetime := types.Time{Time: time.Date(2025, 6, 30, 1, 0, 0, 0, tzAsiaJakarta)}

	tests := []struct {
		name      string
		request   partitionModel.ReorganizeRangePartitionRequest
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Invalid partition value",
			request: partitionModel.ReorganizeRangePartitionRequest{
				ReorganizePartitionConfig: config.RangeTablePartitionConfig{
					Partitions: []config.RangePartitionConfig{
						{
							PartitionName:     "p0",
							DataOlderThanDays: "ABC",
						},
					},
				},
			},
			setupMock: func() { /* Empty */ },
			wantError: errors.New("invalid value for DATA_OLDER_THAN_DAYS in partition p0: strconv.ParseInt: parsing \"ABC\": invalid syntax"),
		},
		{
			name: "ERROR:Not contain MAXVALUE",
			request: partitionModel.ReorganizeRangePartitionRequest{
				ReorganizePartitionConfig: config.RangeTablePartitionConfig{
					Partitions: []config.RangePartitionConfig{
						{
							PartitionName:     "p0",
							DataOlderThanDays: "1",
						},
					},
				},
			},
			setupMock: func() { /* Empty */ },
			wantError: errors.New("failed reorganize partition: range partition without MAXVALUE"),
		},
		{
			name: "SUCCESS:Single partition",
			request: partitionModel.ReorganizeRangePartitionRequest{
				Datetime: requestDatetime,
				ReorganizePartitionConfig: config.RangeTablePartitionConfig{
					Partitions: []config.RangePartitionConfig{
						{PartitionName: "p0", DataOlderThanDays: "0"},
					},
				},
			},
			setupMock: func() {
				mysql.On("ReorganizeRangePartition", mock.Anything, mock.MatchedBy(func(request mySqlExt.ReorganizeRangePartitionRequest) bool {
					return len(request.Partitions) == 1 &&
						request.Partitions[0].PartitionName == "p0" &&
						len(request.Partitions[0].NewPartitions) == 1 &&
						request.Partitions[0].NewPartitions[0].PartitionValue == `MAXVALUE`
				})).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Multi partition",
			request: partitionModel.ReorganizeRangePartitionRequest{
				Datetime: requestDatetime,
				ReorganizePartitionConfig: config.RangeTablePartitionConfig{
					Partitions: []config.RangePartitionConfig{
						{PartitionName: "p_cold_archive", DataOlderThanDays: "180"},
						{PartitionName: "p_cold_historical", DataOlderThanDays: "30"},
						{PartitionName: "p_warm", DataOlderThanDays: "7"},
						{PartitionName: "p_hot", DataOlderThanDays: "0"},
					},
				},
			},
			setupMock: func() {
				mysql.On("ReorganizeRangePartition", mock.Anything, mock.MatchedBy(func(request mySqlExt.ReorganizeRangePartitionRequest) bool {
					return len(request.Partitions) == 4 &&
						len(request.Partitions[0].NewPartitions) == 1 &&
						len(request.Partitions[1].NewPartitions) == 1 &&
						len(request.Partitions[2].NewPartitions) == 1 &&
						len(request.Partitions[3].NewPartitions) == 1 &&
						request.Partitions[0].NewPartitions[0].PartitionValue == `UNIX_TIMESTAMP('2025-01-02 17:00:00')` &&
						request.Partitions[1].NewPartitions[0].PartitionValue == `UNIX_TIMESTAMP('2025-06-01 17:00:00')` &&
						request.Partitions[2].NewPartitions[0].PartitionValue == `UNIX_TIMESTAMP('2025-06-24 17:00:00')` &&
						request.Partitions[3].NewPartitions[0].PartitionValue == `MAXVALUE`
				})).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, db.ReorganizeDailyRangePartition(t.Context(), test.request))
		})
	}
}

func TestGetPartitionNames(t *testing.T) {
	db := mocks.NewIMySqlExt(t)
	db.On("DBName").Return("testdb")

	repo := New(db)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []string
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("failed get partition names: %w", assert.AnError),
		},
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(sql.ErrNoRows)
			},
			wantError: nil, wantResult: nil,
		},
		{
			name: "SUCCESS:Data found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*[]string) = []string{"p_202510", "p_max"}
				}).Return(nil)
			},
			wantResult: []string{"p_202510", "p_max"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetPartitionNames(t.Context(), "tests", "DESC", 10)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}

func TestReorganizeMonthlyRangePartition(t *testing.T) {
	db := mocks.NewIMySqlExt(t)

	repo := New(db)

	queryPatternNewPartition := mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "ALTER TABLE tests ALGORITHM=INPLACE, LOCK=DEFAULT, REORGANIZE PARTITION p_max INTO")
	})
	queryPatternMergedPartitions := mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "ALTER TABLE tests ALGORITHM=INPLACE, LOCK=DEFAULT, REORGANIZE PARTITION p_202506, p_202505, p_202504 INTO")
	})

	tests := []struct {
		name      string
		request   partitionModel.ReorganizeRangePartitionRequest
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:New partition", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				TableName:             "tests",
				Datetime:              types.Time{Time: time.Date(2025, 10, 31, 01, 15, 0, 0, tzAsiaJakarta)},
				MaxValuePartitionName: "p_max",
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, queryPatternNewPartition).Once().Return(false, assert.AnError)
			},
			wantError: fmt.Errorf("failed reorganize partition: %w", assert.AnError),
		},
		{
			name: "ERROR:Get partition names", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				TableName:                    "tests",
				Datetime:                     types.Time{Time: time.Date(2025, 9, 30, 01, 15, 0, 0, tzAsiaJakarta)},
				LastUnmergedMonthlyPartition: 3,
			},
			setupMock: func() {
				db.On("DBName").Return("testdb")
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "ORDER BY PARTITION_ORDINAL_POSITION DESC LIMIT 9")
					}), "testdb", "tests",
				).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("reorganize monthly partition: %w", fmt.Errorf("failed get partition names: %w", assert.AnError)),
		},
		{
			name: "ERROR:Merging partitions", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				TableName:                    "tests",
				Datetime:                     types.Time{Time: time.Date(2025, 9, 30, 01, 15, 0, 0, tzAsiaJakarta)},
				LastUnmergedMonthlyPartition: 3,
				MaxValuePartitionName:        "p_max",
			},
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "ORDER BY PARTITION_ORDINAL_POSITION DESC LIMIT 9")
					}), "testdb", "tests",
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]string) = []string{"p_2025_q1", "p_202504", "p_202505", "p_202506", "p_202507", "p_202508", "p_202509", "p_max"}
				}).Return(nil)
				db.On("ExecContext", mock.Anything, queryPatternNewPartition).Once().Return(true, nil)
				db.On("ExecContext", mock.Anything, queryPatternMergedPartitions).Once().Return(false, assert.AnError)
			},
			wantError: fmt.Errorf("failed reorganize partition: %w", assert.AnError),
		},
		{
			name: "ERROR:Analyze table", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				TableName:                    "tests",
				Datetime:                     types.Time{Time: time.Date(2025, 9, 30, 01, 15, 0, 0, tzAsiaJakarta)},
				LastUnmergedMonthlyPartition: 3,
				MaxValuePartitionName:        "p_max",
				AnalyzeTable:                 true,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, queryPatternNewPartition).Once().Return(true, nil)
				db.On("ExecContext", mock.Anything, queryPatternMergedPartitions).Once().Return(true, nil)
				db.On("ExecContext", mock.Anything, "ANALYZE TABLE tests;").Once().Return(false, assert.AnError)
			},
			wantError: fmt.Errorf("failed analyze table: %w", assert.AnError),
		},
		{
			name: "SUCCESS", // NOSONAR
			request: partitionModel.ReorganizeRangePartitionRequest{
				TableName:                    "tests",
				Datetime:                     types.Time{Time: time.Date(2025, 11, 30, 01, 15, 0, 0, tzAsiaJakarta)},
				LastUnmergedMonthlyPartition: 3,
				MaxValuePartitionName:        "p_max",
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, queryPatternNewPartition).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.ReorganizeMonthlyRangePartition(t.Context(), test.request))

			db.AssertExpectations(t)
		})
	}
}
