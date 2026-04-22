package study

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

// ReportKeyFilters allows optional filtering when listing unique report keys.
// Use 0 values to omit a filter.
type ReportKeyFilters struct {
	ParticipantID string
	FromTS        int64
	ToTS          int64
}

const (
	idxReportsParticipantID             = "participantID_1"
	idxReportsTimestamp                 = "timestamp_1"
	idxReportsParticipantIDKeyTimestamp = "participantID_1_key_1_timestamp_1"
)

var defaultReportIndexNames = []string{
	idxReportsParticipantID,
	idxReportsTimestamp,
	idxReportsParticipantIDKeyTimestamp,
}

var indexesForReportsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "participantID", Value: 1},
		},
		Options: options.Index().SetName(idxReportsParticipantID),
	},
	{
		Keys: bson.D{
			{Key: "timestamp", Value: 1},
		},
		Options: options.Index().SetName(idxReportsTimestamp),
	},
	{
		Keys: bson.D{
			{Key: "participantID", Value: 1},
			{Key: "key", Value: 1},
			{Key: "timestamp", Value: 1},
		},
		Options: options.Index().SetName(idxReportsParticipantIDKeyTimestamp),
	},
}

func (dbService *StudyDBService) DropIndexForReportsCollection(instanceID string, studyKey string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionReports(instanceID, studyKey)

	if dropAll {
		err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for reports", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		}
	} else {
		for _, indexName := range defaultReportIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for reports collection")
				continue
			}
			err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for reports", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForReportsCollection(instanceID string, studyKey string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionReports(instanceID, studyKey)
	_, err := collection.Indexes().CreateMany(ctx, indexesForReportsCollection)
	if err != nil {
		slog.Error("Error creating index for reports", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
	}
}

func (dbService *StudyDBService) SaveReport(instanceID string, studyKey string, report studyTypes.Report) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionReports(instanceID, studyKey).InsertOne(ctx, report)
	return err
}

// get report by id
func (dbService *StudyDBService) GetReportByID(instanceID string, studyKey string, reportID string) (report studyTypes.Report, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := bson.ObjectIDFromHex(reportID)
	if err != nil {
		return report, err
	}

	filter := bson.M{
		"_id": _id,
	}

	err = dbService.collectionReports(instanceID, studyKey).FindOne(ctx, filter).Decode(&report)
	return report, err
}

type UpdateParticipantReportMode string

const (
	UpdateParticipantReportModeAppend  UpdateParticipantReportMode = "append"
	UpdateParticipantReportModeReplace UpdateParticipantReportMode = "replace"
)

// update report data
func (dbService *StudyDBService) UpdateReportData(
	instanceID string,
	studyKey string,
	reportID string,
	participantID string,
	data []studyTypes.ReportData,
	mode UpdateParticipantReportMode,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := bson.ObjectIDFromHex(reportID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": _id, "participantID": participantID}
	update := bson.M{}
	switch mode {
	case UpdateParticipantReportModeAppend:
		update["$push"] = bson.M{"data": bson.M{"$each": data}}
		update["$set"] = bson.M{"modifiedAt": time.Now()}
	case UpdateParticipantReportModeReplace:
		update["$set"] = bson.M{"data": data, "modifiedAt": time.Now()}
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}

	res, err := dbService.collectionReports(instanceID, studyKey).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("report not found, does not belong to participant or could not be updated")
	}
	return nil
}

var reportSortOnTimestamp = bson.D{
	{Key: "timestamp", Value: -1},
}

// get report count for query
func (dbService *StudyDBService) GetReportCountForQuery(instanceID string, studyKey string, filter bson.M) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	return dbService.collectionReports(instanceID, studyKey).CountDocuments(ctx, filter)
}

func (dbService *StudyDBService) UpdateParticipantIDonReports(instanceID string, studyKey string, oldID string, newID string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if oldID == "" || newID == "" {
		return 0, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": oldID}
	update := bson.M{"$set": bson.M{"participantID": newID}}

	res, err := dbService.collectionReports(instanceID, studyKey).UpdateMany(ctx, filter, update)
	return res.ModifiedCount, err
}

// get reports for query with pagination
func (dbService *StudyDBService) GetReports(instanceID string, studyKey string, filter bson.M, page int64, limit int64) (reports []studyTypes.Report, paginationInfo *PaginationInfos, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	totalCount, err := dbService.GetReportCountForQuery(instanceID, studyKey, filter)
	if err != nil {
		return reports, nil, err
	}

	paginationInfo = prepPaginationInfos(
		totalCount,
		page,
		limit,
	)

	skip := (paginationInfo.CurrentPage - 1) * paginationInfo.PageSize

	opts := options.Find()
	opts.SetSort(reportSortOnTimestamp)
	opts.SetSkip(skip)
	opts.SetLimit(paginationInfo.PageSize)

	cursor, err := dbService.collectionReports(instanceID, studyKey).Find(ctx, filter, opts)
	if err != nil {
		return reports, nil, err
	}

	defer cursor.Close(ctx)

	err = cursor.All(ctx, &reports)
	return reports, paginationInfo, err
}

// iterate over reports for query
func (dbService *StudyDBService) FindAndExecuteOnReports(
	ctx context.Context,
	instanceID string, studyKey string,
	filter bson.M,
	returnOnErr bool,
	fn func(instanceID string, studyKey string, report studyTypes.Report, args ...interface{}) error,
	args ...interface{},
) error {
	opts := options.Find().SetSort(reportSortOnTimestamp)

	cursor, err := dbService.collectionReports(instanceID, studyKey).Find(ctx, filter, opts)
	if err != nil {
		return err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var report studyTypes.Report
		if err = cursor.Decode(&report); err != nil {
			slog.Error("Error while decoding report", slog.String("error", err.Error()))
			continue
		}

		if err = fn(instanceID, studyKey, report, args...); err != nil {
			slog.Error("Error executing function on report", slog.String("reportID", report.ID.Hex()), slog.String("error", err.Error()))
			if returnOnErr {
				return err
			}
			continue
		}
	}
	return nil
}

// GetUniqueReportKeysForStudy returns the distinct report keys within a study.
// Optional filters:
// - participantID: if non-empty, limits to reports from the specified participant
// - fromTS/toTS: if >0, applies inclusive timestamp range filters (unix seconds)
func (dbService *StudyDBService) GetUniqueReportKeysForStudy(
	instanceID string,
	studyKey string,
	filters *ReportKeyFilters,
) ([]string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if filters != nil {
		if filters.ParticipantID != "" {
			filter["participantID"] = filters.ParticipantID
		}

		tsFilter := bson.M{}
		if filters.FromTS > 0 {
			tsFilter["$gte"] = filters.FromTS
		}
		if filters.ToTS > 0 {
			tsFilter["$lte"] = filters.ToTS
		}
		if len(tsFilter) > 0 {
			filter["timestamp"] = tsFilter
		}
	}

	var keys []string
	if err := dbService.collectionReports(instanceID, studyKey).Distinct(ctx, "key", filter).Decode(&keys); err != nil {
		return nil, err
	}
	return keys, nil
}
