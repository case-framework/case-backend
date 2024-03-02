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
	COLLECTION_NAME_STUDY_RULES                   = "studyRules"
	COLLECTION_NAME_SUFFIX_SURVEYS                = "surveys"
	COLLECTION_NAME_SUFFIX_RESPONSES              = "surveyResponses"
	COLLECTION_NAME_SUFFIX_PARTICIPANTS           = "participants"
	COLLECTION_NAME_SUFFIX_CONFIDENTIAL_RESPONSES = "confidentialResponses"
	COLLECTION_NAME_SUFFIX_REPORTS                = "reports"
	COLLECTION_NAME_SUFFIX_FILES                  = "participantFiles"
	COLLECTION_NAME_SUFFIX_RESEARCHER_MESSAGES    = "researcherMessages"
	COLLECTION_NAME_TASK_QUEUE                    = "taskQueue"
)

const (
	REMOVE_TASK_FROM_QUEUE_AFTER = 60 * 60 * 24 * 2 // 2 days
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

	if err := studyDBSc.ensureIndexes(); err != nil {
		slog.Error("Error ensuring indexes for study DB: ", err)
	}

	return studyDBSc, nil
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
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(studyKey + "_" + COLLECTION_NAME_SUFFIX_SURVEYS)
}

func (dbService *StudyDBService) collectionResponses(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(studyKey + "_" + COLLECTION_NAME_SUFFIX_RESPONSES)
}

func (dbService *StudyDBService) collectionParticipants(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(studyKey + "_" + COLLECTION_NAME_SUFFIX_PARTICIPANTS)
}

func (dbService *StudyDBService) collectionConfidentialResponses(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(studyKey + "_" + COLLECTION_NAME_SUFFIX_CONFIDENTIAL_RESPONSES)
}

func (dbService *StudyDBService) collectionReports(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(studyKey + "_" + COLLECTION_NAME_SUFFIX_REPORTS)
}

func (dbService *StudyDBService) collectionFiles(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(studyKey + "_" + COLLECTION_NAME_SUFFIX_FILES)
}

func (dbService *StudyDBService) collectionResearcherMessages(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(studyKey + "_" + COLLECTION_NAME_SUFFIX_RESEARCHER_MESSAGES)
}

func (dbService *StudyDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

func (dbService *StudyDBService) ensureIndexes() error {
	slog.Debug("Ensuring indexes for study DB")
	for _, instanceID := range dbService.InstanceIDs {
		ctx, cancel := dbService.getContext()
		defer cancel()

		// task queue: auto delete on creation date
		_, err := dbService.collectionTaskQueue(instanceID).Indexes().CreateOne(
			ctx,
			mongo.IndexModel{
				Keys:    bson.D{{Key: "createdAt", Value: 1}},
				Options: options.Index().SetExpireAfterSeconds(REMOVE_TASK_FROM_QUEUE_AFTER),
			},
		)
		if err != nil {
			slog.Error("Error creating index for createdAt in userDB.sessions: ", err)
		}

		// index on studyInfos
		err = dbService.createIndexForStudyInfosCollection(instanceID)
		if err != nil {
			slog.Error("Error creating index for studyInfos: ", err)
		}

		//fetch studyKeys from studyInfos
		studies, err := dbService.GetStudies(instanceID, "", true)
		if err != nil {
			slog.Error("Error fetching studies", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			return err
		}

		// index on studyRules
		err = dbService.CreateIndexForStudyRulesCollection(instanceID)
		if err != nil {
			slog.Error("Error creating index for studyRules: ", err)
		}

		for _, study := range studies {
			studyKey := study.Key

			// index on surveys
			err = dbService.CreateIndexForSurveyCollection(instanceID, studyKey)
			if err != nil {
				slog.Error("Error creating index for surveys: ", slog.String("error", err.Error()))
			}

			// index on participants
			err = dbService.CreateIndexForParticipantsCollection(instanceID, studyKey)
			if err != nil {
				slog.Error("Error creating index for participants: ", slog.String("error", err.Error()))
			}

			// index on responses
			err = dbService.CreateIndexForResponsesCollection(instanceID, studyKey)
			if err != nil {
				slog.Error("Error creating index for responses: ", slog.String("error", err.Error()))
			}

			// index on reports
			err = dbService.CreateIndexForReportsCollection(instanceID, studyKey)
			if err != nil {
				slog.Error("Error creating index for reports: ", slog.String("error", err.Error()))
			}
		}

	}
	return nil
}
