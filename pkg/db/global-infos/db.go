package globalinfos

import (
	"context"
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collection names
const (
	COLLECTION_NAME_TEMPTOKENS = "temp-tokens"
)

type GlobalInfosDBService struct {
	DBClient        *mongo.Client
	timeout         int
	noCursorTimeout bool
	DBNamePrefix    string
	InstanceIDs     []string
}

func NewGlobalInfosDBService(configs db.DBConfig) (*GlobalInfosDBService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	defer cancel()

	dbClient, err := mongo.Connect(ctx,
		options.Client().ApplyURI(configs.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(configs.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(configs.MaxPoolSize),
	)

	if err != nil {
		return nil, err
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()

	if err != nil {
		return nil, err
	}

	giDBSc := &GlobalInfosDBService{
		DBClient:        dbClient,
		timeout:         configs.Timeout,
		noCursorTimeout: configs.NoCursorTimeout,
		DBNamePrefix:    configs.DBNamePrefix,
		InstanceIDs:     configs.InstanceIDs,
	}

	giDBSc.ensureIndexes()
	return giDBSc, nil
}

func (dbService *GlobalInfosDBService) getDBName() string {
	return dbService.DBNamePrefix + "global-infos"
}

func (dbService *GlobalInfosDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

func (dbService *GlobalInfosDBService) collectionTemptokens() *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName()).Collection(COLLECTION_NAME_TEMPTOKENS)
}

func (dbService *GlobalInfosDBService) ensureIndexes() {
	slog.Debug("Ensuring indexes for global infos DB")

	err := dbService.CreateIndexForTemptokens()
	if err != nil {
		slog.Debug("Error creating indexes for temp tokens: ", slog.String("error", err.Error()))
	}

}
