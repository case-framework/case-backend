package study

import (
	"errors"

	studytypes "github.com/case-framework/case-backend/pkg/study/study"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbService *StudyDBService) FindConfidentialResponses(instanceID string, studyKey string, participantID string, key string) (responses []studytypes.SurveyResponse, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if participantID == "" {
		return responses, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": participantID}
	if key != "" {
		filter["key"] = key
	}

	cur, err := dbService.collectionConfidentialResponses(instanceID, studyKey).Find(
		ctx,
		filter,
		nil,
	)

	if err != nil {
		return responses, err
	}
	defer cur.Close(ctx)

	responses = []studytypes.SurveyResponse{}
	for cur.Next(ctx) {
		var result studytypes.SurveyResponse
		err := cur.Decode(&result)
		if err != nil {
			return responses, err
		}

		responses = append(responses, result)
	}
	if err := cur.Err(); err != nil {
		return responses, err
	}

	return responses, nil
}
