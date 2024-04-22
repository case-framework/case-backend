package study

import "go.mongodb.org/mongo-driver/bson"

func (dbService *StudyDBService) AddConfidentialIDMapEntry(instanceID, confidentialID, profileID, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	entry := bson.M{
		"confidentialID": confidentialID,
		"profileID":      profileID,
		"studyKey":       studyKey,
	}

	_, err := dbService.collectionConfidentialIDMap(instanceID).InsertOne(ctx, entry)
	return err
}

func (dbService *StudyDBService) GetProfileIDFromConfidentialID(instanceID, confidentialID, studyKey string) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"confidentialID": confidentialID,
		"studyKey":       studyKey,
	}

	var result struct {
		ProfileID string `bson:"profileID"`
	}
	err := dbService.collectionConfidentialIDMap(instanceID).FindOne(ctx, filter).Decode(&result)
	return result.ProfileID, err
}

func (dbService *StudyDBService) RemoveConfidentialIDMapEntriesForStudy(instanceID, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionConfidentialIDMap(instanceID).DeleteMany(ctx, bson.M{"studyKey": studyKey})
	return err
}

func (dbService *StudyDBService) RemoveConfidentialIDMapEntriesForProfile(instanceID, profileID, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionConfidentialIDMap(instanceID).DeleteMany(ctx, bson.M{"profileID": profileID, "studyKey": studyKey})
	return err
}
