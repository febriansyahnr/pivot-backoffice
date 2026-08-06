package activityRepository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	activityRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/activity"
	mongoDbMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mongoDbExt"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userID = uuid.NewString()

func TestCreate(t *testing.T) {
	activity := activityModel.Activity{
		ID:          uuid.NewString(),
		MerchantID:  uuid.NewString(),
		UserID:      &userID,
		Tag:         constant.TagAccount,
		Activity:    constant.ActivityUserLogin,
		ServiceName: "DoLogin",
		Parameter: &map[string]any{
			"email":    "jay@paper.id",
			"password": "*****",
		},
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	testCase := []struct {
		name      string
		mockSetup func(mongoDbMock *mongoDbMocks.IMongoDbExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Insert Activity Logs to MongoDB",
			mockSetup: func(mongoDbMock *mongoDbMocks.IMongoDbExt) {
				mongoDbMock.On(
					"InsertOne",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*activityModel.Activity"),
				).Return(&mongo.InsertOneResult{}, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: Failure insert Activity Logs to MongoDB",
			mockSetup: func(mongoDbMock *mongoDbMocks.IMongoDbExt) {
				mongoDbMock.On(
					"InsertOne",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*activityModel.Activity"),
				).Return(nil, errors.New("insert error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMongoDb := mongoDbMocks.NewIMongoDbExt(t)
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			tc.mockSetup(mockMongoDb)

			factory := activityRepository.ActivityRepository{
				Mongo: mockMongoDb,
				Mysql: mockMysql,
			}
			repo := factory.CreateRepository(activityRepository.MongoDBType)
			ctx := context.Background()
			err := repo.Create(ctx, &activity)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMongoDb.AssertExpectations(t)
		})
	}
}

func TestGetList(t *testing.T) {
	merchantID := uuid.NewString()
	testCase := []struct {
		name      string
		mockSetup func(mongoDbMock *mongoDbMocks.IMongoDbExt, mockCursor *mongoDbMocks.ICursorExt)
		filter    activityModel.ActivityFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mongoDbMock *mongoDbMocks.IMongoDbExt, mockCursor *mongoDbMocks.ICursorExt) {
				mockCursor.On("Next", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(true).Times(1)
				mockCursor.On("Next", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(false)
				mockCursor.On("Close", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(nil)
				mockCursor.On("Decode", mock.Anything).Return(nil)

				mongoDbMock.On(
					"Find",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(mockCursor, nil)

				mongoDbMock.On(
					"CountDocument",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(int64(1), nil)
			},
			filter:  activityModel.ActivityFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with merchantID filter",
			mockSetup: func(mongoDbMock *mongoDbMocks.IMongoDbExt, mockCursor *mongoDbMocks.ICursorExt) {
				mockCursor.On("Next", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(false)
				mockCursor.On("Close", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(nil)

				mongoDbMock.On(
					"Find",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(mockCursor, nil)

				mongoDbMock.On(
					"CountDocument",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(int64(1), nil)
			},
			filter: activityModel.ActivityFilterRequest{
				MerchantID: &merchantID,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with created_at filter",
			mockSetup: func(mongoDbMock *mongoDbMocks.IMongoDbExt, mockCursor *mongoDbMocks.ICursorExt) {
				mockCursor.On("Next", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(false)
				mockCursor.On("Close", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(nil)

				mongoDbMock.On(
					"Find",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(mockCursor, nil)

				mongoDbMock.On(
					"CountDocument",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(int64(1), nil)
			},
			filter: activityModel.ActivityFilterRequest{
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get collection",
			mockSetup: func(mongoDbMock *mongoDbMocks.IMongoDbExt, mockCursor *mongoDbMocks.ICursorExt) {
				mongoDbMock.On(
					"Find",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("invalid collection name"))
			},
			filter:  activityModel.ActivityFilterRequest{},
			wantErr: true,
		},
		{
			name: "FAILED: Get List on error retrieving data",
			mockSetup: func(mongoDbMock *mongoDbMocks.IMongoDbExt, mockCursor *mongoDbMocks.ICursorExt) {
				mockCursor.On("Next", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(true)
				mockCursor.On("Close", mock.AnythingOfType(constant.MockTypeValueContextReference)).Return(nil)
				mockCursor.On("Decode", mock.Anything).Return(errors.New("invalid decode data"))

				mongoDbMock.On(
					"Find",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
				).Return(mockCursor, nil)
			},
			filter:  activityModel.ActivityFilterRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMongoDb := mongoDbMocks.NewIMongoDbExt(t)
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockCursor := mongoDbMocks.NewICursorExt(t)
			tc.mockSetup(mockMongoDb, mockCursor)

			factory := activityRepository.ActivityRepository{
				Mongo: mockMongoDb,
				Mysql: mockMysql,
			}
			repo := factory.CreateRepository(activityRepository.MongoDBType)
			ctx := context.Background()
			_, err := repo.GetList(ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMongoDb.AssertExpectations(t)
		})
	}
}

// TestMongoDBRepository_FindLastMerchantActivityDate tests specifically the MongoDB implementation
func TestMongoDBRepository_FindLastMerchantActivityDate(t *testing.T) {
	merchantID := uuid.NewString()
	// Truncate time to milliseconds to avoid precision issues
	expectedTime := time.Now().UTC().Truncate(time.Millisecond)

	// Create a sample activity for decoding
	sampleActivity := activityModel.Activity{
		ID:          uuid.NewString(),
		MerchantID:  merchantID,
		UserID:      &userID,
		Tag:         constant.TagAccount,
		Activity:    constant.ActivityUserLogin,
		ServiceName: "DoLogin",
		Parameter: &map[string]any{
			"email": "test@example.com",
		},
		CreatedAt: expectedTime,
		UpdatedAt: expectedTime,
	}

	testCases := []struct {
		name      string
		setupMock func(mockMongo *mongoDbMocks.IMongoDbExt, singleResult *mongo.SingleResult)
		wantErr   bool
		expected  time.Time
	}{
		{
			name: "SUCCESS: Find last merchant activity date",
			setupMock: func(mockMongo *mongoDbMocks.IMongoDbExt, singleResult *mongo.SingleResult) {
				mockMongo.On(
					"FindOne",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					activityModel.COLLECTION_NAME,
					mock.MatchedBy(func(filter interface{}) bool {
						// Verify filter contains correct merchant_id
						filterM, ok := filter.(bson.M)
						return ok && filterM["merchant_id"] == merchantID
					}),
					mock.AnythingOfType("*options.FindOneOptions"),
				).Return(singleResult)
			},
			wantErr:  false,
			expected: expectedTime,
		},
		{
			name: "ERROR: No activity found",
			setupMock: func(mockMongo *mongoDbMocks.IMongoDbExt, singleResult *mongo.SingleResult) {
				mockMongo.On(
					"FindOne",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					activityModel.COLLECTION_NAME,
					mock.Anything,
					mock.Anything,
				).Return(mongo.NewSingleResultFromDocument(nil, errors.New("document not found"), nil))
			},
			wantErr:  true,
			expected: time.Time{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockMongo := mongoDbMocks.NewIMongoDbExt(t)
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			// For success case, create a mock that will decode our activity
			var singleResult *mongo.SingleResult
			if !tc.wantErr {
				// Marshal the activity to BSON for successful decoding
				activityData, _ := bson.Marshal(sampleActivity)
				singleResult = mongo.NewSingleResultFromDocument(activityData, nil, nil)
			}

			tc.setupMock(mockMongo, singleResult)

			// Create repository using the factory pattern that's already established
			factory := activityRepository.ActivityRepository{
				Mongo: mockMongo,
				Mysql: mockMysql,
			}
			repo := factory.CreateRepository(activityRepository.MongoDBType)

			// Call the method
			result, err := repo.FindLastMerchantActivityDate(context.Background(), merchantID)

			// Verify results
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expected, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			// Verify expectations
			mockMongo.AssertExpectations(t)
		})
	}
}
