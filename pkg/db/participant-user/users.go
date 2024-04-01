package participantuser

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (dbService *ParticipantUserDBService) CreateIndexForParticipantUsers(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionParticipantUsers(instanceID).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "timestamps.markedForDeletion", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "account.accountID", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "timestamps.createdAt", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "account.accountConfirmedAt", Value: 1},
					{Key: "timestamps.createdAt", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "contactPreferences.receiveWeeklyMessageDayOfWeek", Value: 1},
				},
			},
		},
	)
	return err
}
