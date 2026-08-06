package tablePartitionExt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"go.uber.org/zap"
)

// CreateDayRangePartition will return error when failed to validate the config and failed to create the partition
func (p *partitionTable) CreateDayRangePartition(ctx context.Context, cfg PartitionConfig) error {
	ctx, segment := otelTracer.Start(ctx, "pkg/tablePartitionExt/CreateDayRangePartition")
	defer segment.End()

	if err := p.validateConfig(cfg); err != nil {
		return err
	}

	totalPartition, err := p.getTotalPartitions(ctx, cfg)
	if err != nil {
		return err
	}

	if totalPartition == 0 {
		if err = p.initDayRangePartition(ctx, cfg); err != nil {
			return err
		}

		cfg.TotalPartition -= 1
		cfg.StartedAt = cfg.StartedAt.AddDate(0, 0, 1)
	}

	return p.createDayPartitions(ctx, cfg)
}

// validateConfig will check the config and return error when config has invalid value
func (p *partitionTable) validateConfig(cfg PartitionConfig) error {
	if cfg.TableName == "" {
		return errors.New("table name is required")
	}

	if cfg.TotalPartition <= 0 {
		return errors.New("total partition is required")
	}

	if cfg.StartedAt.IsZero() {
		return errors.New("start time is required")
	}

	return nil
}

// getTotalPartitions will return total partition in the table before we initiate the first partition
func (p *partitionTable) getTotalPartitions(ctx context.Context, cfg PartitionConfig) (int, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/tablePartitionExt/getTotalPartitions")
	defer segment.End()

	var (
		totalPartition           int
		queryGetExistingParition = `
		SELECT COUNT(PARTITION_NAME) 
		FROM information_schema.PARTITIONS 
		WHERE 
			TABLE_SCHEMA = ? AND 
			TABLE_NAME = ? AND 
			PARTITION_NAME IS NOT NULL AND 
			PARTITION_NAME != "";`
	)

	err := p.db.GetContext(ctx, &totalPartition, queryGetExistingParition, p.tableSchema, cfg.TableName)
	if err != nil {
		return 0, err
	}

	return totalPartition, nil
}

// isPartitionExist return boolean value to make sure the partition exist or not
func (p *partitionTable) isPartitionExist(ctx context.Context, cfg PartitionConfig, partitionName string) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/tablePartitionExt/isPartitionExist")
	defer segment.End()

	var (
		isExist                  bool
		queryGetExistingParition = `
		SELECT COUNT(PARTITION_NAME) >  0
		FROM information_schema.PARTITIONS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND PARTITION_NAME = ?`
	)

	err := p.db.GetContext(ctx, &isExist, queryGetExistingParition, p.tableSchema, cfg.TableName, partitionName)
	if err != nil {
		return false, err
	}

	return isExist, nil
}

// initPartition will create new Day Range Partition
// when the table have precise timestamp, then it will create partition based on the floor of the timestamp
func (p *partitionTable) initDayRangePartition(ctx context.Context, cfg PartitionConfig) error {
	ctx, segment := otelTracer.Start(ctx, "pkg/tablePartitionExt/initDayRangePartition")
	defer segment.End()

	partitionName := fmt.Sprintf("p%s", cfg.StartedAt.Format(constant.TablePartitionFormat))
	definition := fmt.Sprintf("PARTITION %s VALUES LESS THAN (UNIX_TIMESTAMP('%s'))", partitionName, cfg.StartedAt.AddDate(0, 0, 1).Format("2006-01-02"))
	alterStatement := fmt.Sprintf("ALTER TABLE %s PARTITION BY RANGE (UNIX_TIMESTAMP(%s)) (%s)", cfg.TableName, cfg.Parameter, definition)

	// to resolve precise timestamp issue
	// we need to floor the timestamp to make sure the partition is created based on the day
	// ref: https://bugs.mysql.com/bug.php?id=112181
	if cfg.IsPreciseTimestamp {
		alterStatement = fmt.Sprintf("ALTER TABLE %s PARTITION BY RANGE (FLOOR(UNIX_TIMESTAMP(%s))) (%s)", cfg.TableName, cfg.Parameter, definition)
	}

	if _, err := p.db.ExecContext(ctx, alterStatement); err != nil {
		p.log.Error(ctx, "error initializing table partitions", zap.Error(err))
		return err
	}

	return nil
}

// initDefaultPartition will create new default partition when data was doesn't have partition matched
// when its already exists, then it will return nil
// this default max value should be in the end of the partition to handle incoming data that don't matched with any partition
// cannot be implemented right now
// TODO: create max partition to handle missing partition and escape from error
func (p *partitionTable) ManageDefaultMaxPartition(ctx context.Context, cfg PartitionConfig) error {
	return nil
}

// createDayPartitions will create new n Day Range Partition based on the total partition that should be created
// it will return error when failed to create the partition
// it will check the latest partition and create the missing partition based on the latest partition
func (p *partitionTable) createDayPartitions(ctx context.Context, cfg PartitionConfig) error {
	ctx, segment := otelTracer.Start(ctx, "pkg/tablePartitionExt/createDayPartitions")
	defer segment.End()

	var (
		statements []string
	)

	if cfg.TotalPartition == 0 {
		p.log.Info(ctx, "no new partitions to add, skipping")
		return nil
	}

	missingPartitions, err := p.createMissingPartitionRangeDefinition(ctx, cfg)
	if err != nil {
		return err
	}

	if len(missingPartitions) > 0 {
		statements = append(statements, missingPartitions...)
	}

	for cfg.TotalPartition > 0 {
		partitionName := fmt.Sprintf("p%s", cfg.StartedAt.Format(constant.TablePartitionFormat))

		isExist, err := p.isPartitionExist(ctx, cfg, partitionName)
		if err != nil {
			return err
		}

		cfg.TotalPartition -= 1
		cfg.StartedAt = cfg.StartedAt.AddDate(0, 0, 1)

		if isExist {
			continue
		}

		lessThanDate := cfg.StartedAt.Format("2006-01-02")
		definition := fmt.Sprintf("PARTITION %s VALUES LESS THAN (UNIX_TIMESTAMP('%s'))", partitionName, lessThanDate)
		statements = append(statements, definition)
	}

	if len(statements) > 0 {
		alterTableAddQuery := fmt.Sprintf("ALTER TABLE %s ADD PARTITION (%s)", cfg.TableName, strings.Join(statements, ", "))
		if _, err := p.db.ExecContext(ctx, alterTableAddQuery); err != nil && !strings.Contains(err.Error(), "Duplicate partition name") {
			p.log.Error(ctx, "error when adding new partitions", zap.Error(err))
			return err
		}

		return nil
	}

	p.log.Info(ctx, "no new partitions to add, skipping")
	return nil
}

// createMissingPartitionRangeDefinition will create missing partition based on the latest partition
// it will return error when failed to get the latest partition
func (p *partitionTable) createMissingPartitionRangeDefinition(ctx context.Context, cfg PartitionConfig) ([]string, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/tablePartitionExt/createMissingPartitionRangeDefinition")
	defer segment.End()

	var (
		missingPartition = []string{}
		// we need to start from the beginning of the day
		// because the comparison is based on the day, converted from partition name into time -> it will always be the beginning of the day
		newPartitionDate = time.Date(cfg.StartedAt.Year(), cfg.StartedAt.Month(), cfg.StartedAt.Day(), 0, 0, 0, 0, time.UTC)
	)

	latestPartition, err := p.getLatestPartitionName(ctx, cfg)
	if err != nil {
		p.log.Error(ctx, "err-get-latest-partition", zap.Error(err), zap.String("date", cfg.StartedAt.Format("2006-01-02")))
		return nil, err
	}

	if latestPartition == "" {
		return missingPartition, nil
	}

	latestDatePartition, _ := time.Parse(constant.TablePartitionFormat, latestPartition[1:]) // always start from the beginning of the day
	latestDatePartition = latestDatePartition.AddDate(0, 0, 1)

	for newPartitionDate.After(latestDatePartition) {
		partitionName := fmt.Sprintf("p%s", latestDatePartition.Format(constant.TablePartitionFormat))
		lessThanDate := latestDatePartition.AddDate(0, 0, 1).Format("2006-01-02")
		definition := fmt.Sprintf("PARTITION %s VALUES LESS THAN (UNIX_TIMESTAMP('%s'))", partitionName, lessThanDate)
		missingPartition = append(missingPartition, definition)
		latestDatePartition = latestDatePartition.AddDate(0, 0, 1)
	}

	return missingPartition, nil
}

// getLatestPartitionName will return the latest partition name in the table
func (p *partitionTable) getLatestPartitionName(ctx context.Context, cfg PartitionConfig) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/tablePartitionExt/getLatestPartition")
	defer segment.End()

	var (
		partitionName            string
		queryGetExistingParition = `
		SELECT PARTITION_NAME
		FROM information_schema.PARTITIONS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY PARTITION_ORDINAL_POSITION DESC
		LIMIT 1;`
	)

	err := p.db.GetContext(ctx, &partitionName, queryGetExistingParition, p.tableSchema, cfg.TableName)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	return partitionName, nil
}
