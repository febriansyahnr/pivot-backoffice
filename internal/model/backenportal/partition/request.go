package partitionModel

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"
)

type ReorganizeRangePartitionRequest struct {
	TableName                    string
	MaxValuePartitionName        string
	Datetime                     types.Time // Must be in UTC+7 time zone (Asia/Jakarta)
	LastUnmergedMonthlyPartition int
	AnalyzeTable                 bool
	// Deprecated
	ReorganizePartitionConfig config.RangeTablePartitionConfig
}
