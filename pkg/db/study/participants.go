package study

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

const (
	FALLBACK_PAGE_SIZE = 10
)

func (dbService *StudyDBService) CreateIndexForParticipantsCollection(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionParticipants(instanceID, studyKey)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "participantID", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "studyStatus", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "messages.scheduledFor", Value: 1},
				{Key: "studyStatus", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "messages.scheduledFor", Value: 1},
			},
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// get participant by id
func (dbService *StudyDBService) GetParticipantByID(instanceID string, studyKey string, participantID string) (participant studyTypes.Participant, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(participantID)
	if err != nil {
		return participant, err
	}

	filter := bson.M{
		"_id": _id,
	}

	err = dbService.collectionParticipants(instanceID, studyKey).FindOne(ctx, filter).Decode(&participant)
	return participant, err
}

type PaginationInfos struct {
	TotalCount  int64 `json:"totalCount"`
	CurrentPage int64 `json:"currentPage"`
	TotalPages  int64 `json:"totalPages"`
	PageSize    int64 `json:"pageSize"`
}

// get paginated set of participants
func (dbService *StudyDBService) GetParticipants(instanceID string, studyKey string, filter bson.M, sort bson.M, page int64, limit int64) (participants []studyTypes.Participant, paginationInfo *PaginationInfos, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	count, err := dbService.GetParticipantCount(instanceID, studyKey, filter)
	if err != nil {
		return participants, paginationInfo, err
	}

	if limit == 0 {
		limit = FALLBACK_PAGE_SIZE
	}

	paginationInfo = &PaginationInfos{
		PageSize:    limit,
		TotalCount:  count,
		CurrentPage: page,
		TotalPages:  count / limit,
	}

	skip := (page - 1) * limit

	opts := options.Find()
	opts.SetSort(sort)
	opts.SetSkip(skip)
	opts.SetLimit(limit)

	cursor, err := dbService.collectionParticipants(instanceID, studyKey).Find(ctx, filter, opts)
	if err != nil {
		return participants, nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &participants)
	return participants, paginationInfo, err
}

// get participant count for filter
func (dbService *StudyDBService) GetParticipantCount(instanceID string, studyKey string, filter bson.M) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	return dbService.collectionParticipants(instanceID, studyKey).CountDocuments(ctx, filter)
}

// execute function on participants
func (dbService *StudyDBService) FindAndExecuteOnParticipantsStates(
	ctx context.Context,
	instanceID string,
	studyKey string,
	filter bson.M,
	fn func(dbService *StudyDBService, p studyTypes.Participant, instanceID string, studyKey string, args ...interface{}) error,
	args ...interface{},
) error {

	cursor, err := dbService.collectionParticipants(instanceID, studyKey).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var participant studyTypes.Participant
		if err = cursor.Decode(&participant); err != nil {
			return err
		}
		if err = fn(
			dbService,
			participant,
			instanceID,
			studyKey,
			args...,
		); err != nil {
			slog.Error("Error executing function on participant", slog.String("participantID", participant.ParticipantID), slog.String("error", err.Error()))
			continue
		}
	}
	return nil
}

// delete participant
func (dbService *StudyDBService) DeleteParticipantByID(instanceID string, studyKey string, participantID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(participantID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}
	res, err := dbService.collectionParticipants(instanceID, studyKey).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
