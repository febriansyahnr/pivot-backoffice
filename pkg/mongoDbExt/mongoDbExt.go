package mongoDbExt

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/url"
)

type IMongoDbExt interface {
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	InsertOne(ctx context.Context, collectionName string, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	Find(ctx context.Context, collectionName string, filter interface{}, opts ...*options.FindOptions) (ICursorExt, error)
	FindOne(ctx context.Context, collectionName string, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	CountDocument(ctx context.Context, collectionName string, filter interface{}, opts ...*options.CountOptions) (int64, error)
}

type ICursorExt interface {
	Next(ctx context.Context) bool
	Decode(v interface{}) error
	Close(ctx context.Context) error
}

type mongoDbExt struct {
	client   *mongo.Client
	database *mongo.Database
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

func New(config Config) (IMongoDbExt, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/?authSource=%s",
		config.Username,
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		config.DBName,
	)))
	if err != nil {
		return nil, err
	}

	return &mongoDbExt{client: client, database: client.Database(config.DBName)}, nil
}

func (m *mongoDbExt) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *mongoDbExt) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

func (m *mongoDbExt) InsertOne(ctx context.Context, collectionName string, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.database.Collection(collectionName).InsertOne(ctx, document, opts...)
}

func (m *mongoDbExt) Find(ctx context.Context, collectionName string, filter interface{}, opts ...*options.FindOptions) (ICursorExt, error) {
	return m.database.Collection(collectionName).Find(ctx, filter, opts...)
}

func (m *mongoDbExt) FindOne(ctx context.Context, collectionName string, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return m.database.Collection(collectionName).FindOne(ctx, filter, opts...)
}

func (m *mongoDbExt) CountDocument(ctx context.Context, collectionName string, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.database.Collection(collectionName).CountDocuments(ctx, filter, opts...)
}
