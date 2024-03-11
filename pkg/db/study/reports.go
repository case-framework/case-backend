package study

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

func (dbService *StudyDBService) CreateIndexForReportsCollection(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionReports(instanceID, studyKey)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "participantID", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "participantID", Value: 1},
				{Key: "key", Value: 1},
				{Key: "timestamp", Value: 1},
			},
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// get report by id
func (dbService *StudyDBService) GetReportByID(instanceID string, studyKey string, reportID string) (report studyTypes.Report, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return report, err
	}

	filter := bson.M{
		"_id": _id,
	}

	err = dbService.collectionReports(instanceID, studyKey).FindOne(ctx, filter).Decode(&report)
	return report, err
}

var reportSortOnTimestamp = bson.D{
	primitive.E{Key: "timestamp", Value: -1},
}

// get report count for query
func (dbService *StudyDBService) GetReportCountForQuery(instanceID string, studyKey string, filter bson.M) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	return dbService.collectionReports(instanceID, studyKey).CountDocuments(ctx, filter)
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
			continue
		}
	}
	return nil
}
