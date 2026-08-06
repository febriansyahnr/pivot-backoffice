package activityRepository

import (
	"context"
	"math"
	"time"

	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mongoDbExt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBRepository struct {
	mongo mongoDbExt.IMongoDbExt
}

func (r *MongoDBRepository) Create(ctx context.Context, model *activityModel.Activity) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/activity/mongoDbActivity/Create")
	defer segment.End()

	_, err := r.mongo.InsertOne(ctx, activityModel.COLLECTION_NAME, model)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoDBRepository) GetList(
	ctx context.Context,
	filter activityModel.ActivityFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/activity/mongoDbActivity/GetList")
	defer segment.End()

	queryCondition := bson.M{}
	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		queryCondition["created_at"] = bson.M{
			"$gte": filter.StartCreatedAt,
			"$lt":  filter.EndCreatedAt,
		}
	}

	if filter.MerchantID != nil {
		queryCondition["merchant_id"] = *filter.MerchantID
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage
	findOptions := options.Find().SetLimit(perPage).SetSkip(offset)

	csr, err := r.mongo.Find(ctx, activityModel.COLLECTION_NAME, queryCondition, findOptions)
	if err != nil {
		return nil, err
	}
	defer csr.Close(ctx)

	data := make([]activityModel.Activity, 0)
	for csr.Next(ctx) {
		var row activityModel.Activity
		err := csr.Decode(&row)
		if err != nil {
			return nil, err
		}

		data = append(data, row)
	}

	var totalItems int64
	totalItems, _ = r.mongo.CountDocument(ctx, activityModel.COLLECTION_NAME, queryCondition)
	totalPages := int64(math.Ceil(float64(totalItems) / float64(perPage)))

	meta := commonModel.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return &commonModel.PaginationResponse{
		Data: data,
		Meta: meta,
	}, nil
}

// FindLastMerchantActivityDate implements the required method from the IActivityRepository interface
func (r *MongoDBRepository) FindLastMerchantActivityDate(ctx context.Context, merchantId string) (time.Time, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/activity/mongoDbActivity/FindLastMerchantActivityDate")
	defer segment.End()

	// Create filter for finding activities for this merchant
	filter := bson.M{"merchant_id": merchantId}

	// Sort by created_at in descending order to get the most recent activity
	opts := options.FindOne().SetSort(bson.M{"created_at": -1})

	// Find the most recent activity
	var activity activityModel.Activity
	err := r.mongo.FindOne(ctx, activityModel.COLLECTION_NAME, filter, opts).Decode(&activity)
	if err != nil {
		return time.Time{}, err
	}

	return activity.CreatedAt, nil
}
