package study

import (
	"context"
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collection names
const (
	COLLECTION_NAME_STUDY_INFOS                   = "study-infos"
	COLLECTION_NAME_CONFIDENTIAL_ID_MAP           = "confidential-id-map"
	COLLECTION_NAME_STUDY_RULES                   = "studyRules"
	COLLECTION_NAME_SUFFIX_SURVEYS                = "surveys"
	COLLECTION_NAME_SUFFIX_RESPONSES              = "surveyResponses"
	COLLECTION_NAME_SUFFIX_PARTICIPANTS           = "participants"
	COLLECTION_NAME_SUFFIX_CONFIDENTIAL_RESPONSES = "confidentialResponses"
	COLLECTION_NAME_SUFFIX_REPORTS                = "reports"
	COLLECTION_NAME_SUFFIX_FILES                  = "participantFiles"
	COLLECTION_NAME_SUFFIX_RESEARCHER_MESSAGES    = "researcherMessages"
	COLLECTION_NAME_TASK_QUEUE                    = "taskQueue"
	COLLECTION_NAME_STUDY_CODE_LISTS              = "studyCodeLists"
	COLLECTION_NAME_STUDY_COUNTERS                = "studyCounters"
	COLLECTION_NAME_STUDY_VARIABLES               = "studyVariables"
)

type StudyDBService struct {
	DBClient        *mongo.Client
	timeout         int
	noCursorTimeout bool
	DBNamePrefix    string
	InstanceIDs     []string
}

func NewStudyDBService(configs db.DBConfig) (*StudyDBService, error) {
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

	studyDBSc := &StudyDBService{
		DBClient:        dbClient,
		timeout:         configs.Timeout,
		noCursorTimeout: configs.NoCursorTimeout,
		DBNamePrefix:    configs.DBNamePrefix,
		InstanceIDs:     configs.InstanceIDs,
	}

	return studyDBSc, nil
}

func collectionNameWithStudyKeyPrefix(studyKey string, collectionName string) string {
	return studyKey + "_" + collectionName
}

func (dbService *StudyDBService) getDBName(instanceID string) string {
	return dbService.DBNamePrefix + instanceID + "_studyDB"
}

func (dbService *StudyDBService) collectionStudyInfos(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_STUDY_INFOS)
}

func (dbService *StudyDBService) collectionStudyRules(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_STUDY_RULES)
}

func (dbService *StudyDBService) collectionTaskQueue(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_TASK_QUEUE)
}

func (dbService *StudyDBService) collectionSurveys(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_SURVEYS))
}

func (dbService *StudyDBService) collectionResponses(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_RESPONSES))
}

func (dbService *StudyDBService) collectionParticipants(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_PARTICIPANTS))
}

func (dbService *StudyDBService) collectionConfidentialResponses(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_CONFIDENTIAL_RESPONSES))
}

func (dbService *StudyDBService) collectionConfidentialIDMap(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_CONFIDENTIAL_ID_MAP)
}

func (dbService *StudyDBService) collectionReports(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_REPORTS))
}

func (dbService *StudyDBService) collectionFiles(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_FILES))
}

func (dbService *StudyDBService) collectionResearcherMessages(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_RESEARCHER_MESSAGES))
}

func (dbService *StudyDBService) collectionStudyCodeLists(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_STUDY_CODE_LISTS)
}

func (dbService *StudyDBService) collectionStudyCounters(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_STUDY_COUNTERS)
}

func (dbService *StudyDBService) collectionStudyVariables(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_STUDY_VARIABLES)
}

func (dbService *StudyDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

func (dbService *StudyDBService) dropIndexes(all bool) {
	for _, instanceID := range dbService.InstanceIDs {
		start := time.Now()
		slog.Info("Dropping indexes for study DB", slog.String("instanceID", instanceID))

		dbService.DropIndexForConfidentialIDMapCollection(instanceID, all)
		dbService.DropIndexForStudyCodeListsCollection(instanceID, all)
		dbService.DropIndexForStudyCountersCollection(instanceID, all)
		dbService.DropIndexForStudyInfosCollection(instanceID, all)
		dbService.DropIndexForStudyRulesCollection(instanceID, all)
		dbService.DropIndexForTaskQueueCollection(instanceID, all)
		dbService.DropIndexForStudyVariablesCollection(instanceID, all)
		// participant files has no default indexes at the moment
		// researcher messages has no default indexes at the moment

		//fetch studyKeys from studyInfos
		studies, err := dbService.GetStudies(instanceID, "", true)
		if err != nil {
			slog.Error("Error fetching studies", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		for _, study := range studies {
			studyKey := study.Key
			dbService.DropIndexForSurveysCollection(instanceID, studyKey, all)
			dbService.DropIndexForResponsesCollection(instanceID, studyKey, all)
			dbService.DropIndexForConfidentialResponsesCollection(instanceID, studyKey, all)
			dbService.DropIndexForReportsCollection(instanceID, studyKey, all)
			dbService.DropIndexForParticipantsCollection(instanceID, studyKey, all)
		}

		slog.Info("Indexes dropped for study DB", slog.String("instanceID", instanceID), slog.String("duration", time.Since(start).String()))
	}
}

// DropAllIndexes drops all indexes for all instanceIDs and all collections
func (dbService *StudyDBService) DropAllIndexes() {
	dbService.dropIndexes(true)
}

// DropDefaultIndexes drops all default indexes for all instanceIDs and all collections
func (dbService *StudyDBService) DropDefaultIndexes() {
	dbService.dropIndexes(false)
}

// CreateDefaultIndexes creates all default indexes for all instanceIDs and all collections
func (dbService *StudyDBService) CreateDefaultIndexes() {
	for _, instanceID := range dbService.InstanceIDs {
		start := time.Now()
		slog.Info("Creating default indexes for study DB", slog.String("instanceID", instanceID))

		//fetch studyKeys from studyInfos
		studies, err := dbService.GetStudies(instanceID, "", true)
		if err != nil {
			slog.Error("Error fetching studies", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		dbService.CreateDefaultIndexesForConfidentialIDMapCollection(instanceID)
		dbService.CreateDefaultIndexesForStudyCodeListsCollection(instanceID)
		dbService.CreateDefaultIndexesForStudyCountersCollection(instanceID)
		dbService.CreateDefaultIndexesForStudyInfosCollection(instanceID)
		dbService.CreateDefaultIndexesForStudyRulesCollection(instanceID)
		dbService.CreateDefaultIndexesForTaskQueueCollection(instanceID)
		dbService.CreateDefaultIndexesForStudyVariablesCollection(instanceID)
		// participant files has no default indexes at the moment
		// researcher messages has no default indexes at the moment

		for _, study := range studies {
			studyKey := study.Key

			dbService.CreateDefaultIndexesForSurveysCollection(instanceID, studyKey)
			dbService.CreateDefaultIndexesForResponsesCollection(instanceID, studyKey)
			dbService.CreateDefaultIndexesForConfidentialResponsesCollection(instanceID, studyKey)
			dbService.CreateDefaultIndexesForReportsCollection(instanceID, studyKey)
			dbService.CreateDefaultIndexesForParticipantsCollection(instanceID, studyKey)
		}
		slog.Info("Default indexes created for study DB", slog.String("instanceID", instanceID), slog.String("duration", time.Since(start).String()))
	}
}

func (dbService *StudyDBService) GetIndexes() (map[string]map[string][]bson.M, error) {
	results := make(map[string]map[string][]bson.M, len(dbService.InstanceIDs))

	ctx, cancel := dbService.getContext()
	defer cancel()

	for _, instanceID := range dbService.InstanceIDs {
		collectionIndexes := make(map[string][]bson.M)

		var err error
		if collectionIndexes[COLLECTION_NAME_CONFIDENTIAL_ID_MAP], err = db.ListCollectionIndexes(ctx, dbService.collectionConfidentialIDMap(instanceID)); err != nil {
			return nil, err
		}
		if collectionIndexes[COLLECTION_NAME_STUDY_CODE_LISTS], err = db.ListCollectionIndexes(ctx, dbService.collectionStudyCodeLists(instanceID)); err != nil {
			return nil, err
		}
		if collectionIndexes[COLLECTION_NAME_STUDY_COUNTERS], err = db.ListCollectionIndexes(ctx, dbService.collectionStudyCounters(instanceID)); err != nil {
			return nil, err
		}
		if collectionIndexes[COLLECTION_NAME_STUDY_INFOS], err = db.ListCollectionIndexes(ctx, dbService.collectionStudyInfos(instanceID)); err != nil {
			return nil, err
		}
		if collectionIndexes[COLLECTION_NAME_STUDY_RULES], err = db.ListCollectionIndexes(ctx, dbService.collectionStudyRules(instanceID)); err != nil {
			return nil, err
		}
		if collectionIndexes[COLLECTION_NAME_TASK_QUEUE], err = db.ListCollectionIndexes(ctx, dbService.collectionTaskQueue(instanceID)); err != nil {
			return nil, err
		}
		if collectionIndexes[COLLECTION_NAME_STUDY_VARIABLES], err = db.ListCollectionIndexes(ctx, dbService.collectionStudyVariables(instanceID)); err != nil {
			return nil, err
		}

		studies, err := dbService.GetStudies(instanceID, "", true)
		if err != nil {
			return nil, err
		}

		for _, study := range studies {
			studyKey := study.Key

			if collectionIndexes[collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_SURVEYS)], err = db.ListCollectionIndexes(ctx, dbService.collectionSurveys(instanceID, studyKey)); err != nil {
				return nil, err
			}

			if collectionIndexes[collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_RESPONSES)], err = db.ListCollectionIndexes(ctx, dbService.collectionResponses(instanceID, studyKey)); err != nil {
				return nil, err
			}

			if collectionIndexes[collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_CONFIDENTIAL_RESPONSES)], err = db.ListCollectionIndexes(ctx, dbService.collectionConfidentialResponses(instanceID, studyKey)); err != nil {
				return nil, err
			}

			if collectionIndexes[collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_REPORTS)], err = db.ListCollectionIndexes(ctx, dbService.collectionReports(instanceID, studyKey)); err != nil {
				return nil, err
			}

			if collectionIndexes[collectionNameWithStudyKeyPrefix(studyKey, COLLECTION_NAME_SUFFIX_PARTICIPANTS)], err = db.ListCollectionIndexes(ctx, dbService.collectionParticipants(instanceID, studyKey)); err != nil {
				return nil, err
			}
		}

		results[instanceID] = collectionIndexes
	}

	return results, nil
}
