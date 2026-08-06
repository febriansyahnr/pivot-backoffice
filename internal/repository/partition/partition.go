package partitionRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/mySqlExt"
)

const quarterMonths = 3

func (r *repository) GetPartitionNames(ctx context.Context, tableName, ordinalPartitionOrder string, limit int) (pnames []string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/partition/GetPartitionNames")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "INFORMATION_SCHEMA.PARTITIONS")

	query := fmt.Sprintf(`SELECT
		PARTITION_NAME AS name
	FROM INFORMATION_SCHEMA.PARTITIONS
	WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY PARTITION_ORDINAL_POSITION %s LIMIT %d;`, ordinalPartitionOrder, limit)

	if err = r.db.SelectContext(ctx, &pnames, query, r.db.DBName(), tableName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed get partition names: %w", err)
	}
	return pnames, nil
}

// The ReorganizeMonthlyRangePartition function is used to manage range partitions by creating a new partition
// for the following month and merging partitions quarterly based on the provided input.
func (r *repository) ReorganizeMonthlyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/partition/ReorganizeMonthlyRangePartition")
	defer segment.End()

	// The "Datetime" param always has the value of the end of the month.
	ptime := request.Datetime.NextMidnight()

	// The variable "ptime" will always be the first of the month.
	newPartitionName := ptime.Format("p_200601")
	newPartitionValue := ptime.EndOfMonth().NextMidnight().UTC().Format(time.DateTime)

	queries := []string{
		fmt.Sprintf(`ALTER TABLE %s ALGORITHM=INPLACE, LOCK=DEFAULT, REORGANIZE PARTITION %s INTO (
			PARTITION %s VALUES LESS THAN ( UNIX_TIMESTAMP('%s') ),
			PARTITION %s VALUES LESS THAN ( MAXVALUE )
		);`, request.TableName, request.MaxValuePartitionName, newPartitionName, newPartitionValue, request.MaxValuePartitionName),
	}
	if request.Datetime.IsEndOfQuerter() {

		pnames, err := r.GetPartitionNames(ctx, request.TableName, "DESC", (request.LastUnmergedMonthlyPartition + quarterMonths + 3))
		if err != nil {
			return fmt.Errorf("reorganize monthly partition: %w", err)
		}

		dt := ptime
		for range request.LastUnmergedMonthlyPartition + 1 {
			dt = types.Time{Time: dt.AddDate(0, -1, 0)}
		}
		mergedPartitions := make([]string, 0, quarterMonths)
		mergedPartitionName := fmt.Sprintf("%s_q%d", dt.Format("p_2006"), dt.Quarter())
		mergedPartitionValue := dt.EndOfMonth().NextMidnight().UTC().Format(time.DateTime)

		// Looking for the last 3 months (quarter) partition
		for range quarterMonths {
			pname := dt.Format("p_200601")
			if slices.Contains(pnames, pname) {
				mergedPartitions = append(mergedPartitions, pname)
			}
			dt = types.Time{Time: dt.AddDate(0, -1, 0)}
		}

		if len(mergedPartitions) > 0 {
			queries = append(queries, fmt.Sprintf(`ALTER TABLE %s ALGORITHM=INPLACE, LOCK=DEFAULT, REORGANIZE PARTITION %s INTO (
				PARTITION %s VALUES LESS THAN ( UNIX_TIMESTAMP('%s') )
			);`, request.TableName, strings.Join(mergedPartitions, ", "), mergedPartitionName, mergedPartitionValue))
		}
	}

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, request.TableName)

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed reorganize partition: %w", err)
		}
	}
	if request.AnalyzeTable {
		if _, err := r.db.ExecContext(ctx, fmt.Sprintf("ANALYZE TABLE %s;", request.TableName)); err != nil {
			return fmt.Errorf("failed analyze table: %w", err)
		}
	}
	return nil
}

func (r *repository) ReorganizeDailyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/partition/ReorganizeDailyRangePartition")
	defer segment.End()

	reorganizeRequest := mySqlExt.ReorganizeRangePartitionRequest{
		TableName:    request.TableName,
		AnalyzeTable: request.ReorganizePartitionConfig.AnalyzeTable,
	}
	// UTC + 7 time zone (Asia/Jakarta)
	var midnight = time.Date(request.Datetime.Year(), request.Datetime.Month(), request.Datetime.Day(), 23, 59, 59, 0, request.Datetime.Location()).Add(time.Second)

	for _, rop := range request.ReorganizePartitionConfig.Partitions {
		days, err := strconv.ParseInt(rop.DataOlderThanDays, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid value for DATA_OLDER_THAN_DAYS in partition %s: %v", rop.PartitionName, err)
		}
		// UTC + 0 time zone
		var partitionValue = mySqlExt.MaxTimeValue
		if days > 0 {
			partitionValue = midnight.Add(time.Duration(-((days - 1) * 24)) * time.Hour).UTC()
		}
		reorganizeRequest.Partitions = append(reorganizeRequest.Partitions, mySqlExt.ReorganizeRequest{
			PartitionName: rop.PartitionName,
			NewPartitions: []mySqlExt.PartitionRequest{
				{
					PartitionName:  rop.PartitionName,
					PartitionValue: mySqlExt.ToUnixTimestamp(partitionValue),
				},
			},
		})
	}

	containMaxValue := slices.ContainsFunc(reorganizeRequest.Partitions, func(r0 mySqlExt.ReorganizeRequest) bool {
		return slices.ContainsFunc(r0.NewPartitions, func(r1 mySqlExt.PartitionRequest) bool {
			return r1.PartitionValue == mySqlExt.ToUnixTimestamp(mySqlExt.MaxTimeValue)
		})
	})
	if !containMaxValue {
		return errors.New("failed reorganize partition: range partition without MAXVALUE")
	}
	return r.db.ReorganizeRangePartition(ctx, reorganizeRequest)
}
