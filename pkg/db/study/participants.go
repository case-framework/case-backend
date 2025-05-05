package study

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func (dbService *StudyDBService) CreateIndexForParticipantsCollection(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if _, err := dbService.collectionParticipants(instanceID, studyKey).Indexes().DropAll(ctx); err != nil {
		slog.Error("Error dropping indexes for participants", slog.String("error", err.Error()))
	}

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
				{Key: "enteredAt", Value: 1},
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

func (dbService *StudyDBService) SaveParticipantState(instanceID string, studyKey string, pState studyTypes.Participant) (studyTypes.Participant, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"participantID": pState.ParticipantID}
	pState.ModifiedAt = time.Now().Unix()

	upsert := true
	rd := options.After
	options := options.FindOneAndReplaceOptions{
		Upsert:         &upsert,
		ReturnDocument: &rd,
	}
	elem := studyTypes.Participant{}
	err := dbService.collectionParticipants(instanceID, studyKey).FindOneAndReplace(
		ctx, filter, pState, &options,
	).Decode(&elem)
	return elem, err
}

// get participant by id
func (dbService *StudyDBService) GetParticipantByID(instanceID string, studyKey string, participantID string) (participant studyTypes.Participant, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"participantID": participantID,
	}

	err = dbService.collectionParticipants(instanceID, studyKey).FindOne(ctx, filter).Decode(&participant)
	return participant, err
}

// get paginated set of participants
func (dbService *StudyDBService) GetParticipants(instanceID string, studyKey string, filter bson.M, sort bson.M, page int64, limit int64) (participants []studyTypes.Participant, paginationInfo *PaginationInfos, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	count, err := dbService.GetParticipantCount(instanceID, studyKey, filter)
	if err != nil {
		return participants, paginationInfo, err
	}

	paginationInfo = prepPaginationInfos(
		count,
		page,
		limit,
	)

	skip := (paginationInfo.CurrentPage - 1) * paginationInfo.PageSize

	opts := options.Find()
	opts.SetSort(sort)
	opts.SetSkip(skip)
	opts.SetLimit(paginationInfo.PageSize)

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
	sort bson.M,
	returnOnErr bool,
	fn func(dbService *StudyDBService, p studyTypes.Participant, instanceID string, studyKey string, args ...interface{}) error,
	args ...interface{},
) error {
	opts := options.Find()
	opts.SetSort(sort)

	cursor, err := dbService.collectionParticipants(instanceID, studyKey).Find(ctx, filter, opts)
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
			if returnOnErr {
				return err
			}
			continue
		}
	}
	return nil
}

// delete participant
func (dbService *StudyDBService) DeleteParticipantByID(instanceID string, studyKey string, participantID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"participantID": participantID,
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

func (dbService *StudyDBService) DeleteMessagesFromParticipant(instanceID string, studyKey string, participantID string, messageIDs []string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"participantID": participantID}
	update := bson.M{"$pull": bson.M{"messages": bson.M{
		"id": bson.M{"$in": messageIDs},
	}}}
	_, err := dbService.collectionParticipants(instanceID, studyKey).UpdateOne(ctx, filter, update)
	return err
}
