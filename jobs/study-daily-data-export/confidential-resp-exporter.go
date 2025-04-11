package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	studyutils "github.com/case-framework/case-backend/pkg/study/utils"
)

func runConfidentialResponsesExportsForTask(rExpTask ConfidentialResponsesExportTask) {
	// ensure there is a folder path for the source (export_path/instance_id/study_key)
	relativeFolderPath := path.Join(rExpTask.InstanceID, rExpTask.StudyKey)
	exportFolderPathForTask := path.Join(conf.ExportPath, relativeFolderPath)
	if _, err := os.Stat(exportFolderPathForTask); os.IsNotExist(err) {
		// create folder
		err = os.MkdirAll(exportFolderPathForTask, os.ModePerm)
		if err != nil {
			slog.Error("Error creating export path", slog.String("error", err.Error()))
			return
		}
		slog.Info("Created export path", slog.String("path", exportFolderPathForTask))
	}

	// get study
	study, err := studyDBService.GetStudy(rExpTask.InstanceID, rExpTask.StudyKey)
	if err != nil {
		slog.Error("failed to get study", slog.String("error", err.Error()))
		return
	}

	studySecretKey := study.SecretKey
	idMappingMethod := study.Configs.IdMappingMethod
	globalSecret := rExpTask.StudyGlobalSecret

	// export to file (CSV or JSON)
	results := []studyutils.ConfidentialResponsesExportEntry{}

	if err := studyDBService.FindAndExecuteOnConfidentialResponses(
		context.Background(),
		rExpTask.InstanceID,
		rExpTask.StudyKey,
		false,
		func(r studyTypes.SurveyResponse, args ...interface{}) error {
			// compute real participantID (confidentialID -> profileID -> participantID)
			confidentialID := r.ParticipantID
			profileID, err := studyDBService.GetProfileIDFromConfidentialID(rExpTask.InstanceID, confidentialID, rExpTask.StudyKey)
			if err != nil {
				slog.Error("can't find profileID based on confidentialID, you may need to generate profileID lookup", slog.String("error", err.Error()))
				return nil
			}

			pID, err := studyutils.ProfileIDtoParticipantID(profileID, globalSecret, studySecretKey, idMappingMethod)
			if err != nil {
				slog.Error("failed to compute participantID", slog.String("error", err.Error()))
				return nil
			}

			results = append(results, studyutils.PrepConfidentialResponseExport(r, pID, rExpTask.RespKeyFilter)...)
			return nil
		},
	); err != nil {
		slog.Error("failed to execute on confidential responses", slog.String("error", err.Error()))
		return
	}

	// write to file (CSV or JSON)
	if rExpTask.ExportFormat == "csv" {
		err = writeCSV(results, rExpTask.Name, exportFolderPathForTask)
	} else {
		err = writeJSON(results, rExpTask.Name, exportFolderPathForTask)
	}
	if err != nil {
		slog.Error("failed to write to file", slog.String("error", err.Error()))
		return
	}
}

func confidentialResponsesExportFileName(name string, exportFormat string) string {
	parts := []string{
		time.Now().Format("2006-01-02"),
		"confidential-responses",
		name,
		exportFormat,
	}
	return strings.Join(parts, "##") + "." + exportFormat
}

func writeCSV(results []studyutils.ConfidentialResponsesExportEntry, name string, exportFolderPathForTask string) error {
	filename := confidentialResponsesExportFileName(name, "csv")
	file, err := os.Create(filepath.Join(exportFolderPathForTask, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write header
	header := []string{"participantID", "entryID", "responseKey", "value"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, r := range results {
		row := []string{r.ParticipantID, r.EntryID, r.ResponseKey, r.Value}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func writeJSON(results []studyutils.ConfidentialResponsesExportEntry, name string, exportFolderPathForTask string) error {
	filename := confidentialResponsesExportFileName(name, "json")
	file, err := os.Create(filepath.Join(exportFolderPathForTask, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(results)
	if err != nil {
		return err
	}
	return nil
}

func parseSlots(respItem *studyTypes.ResponseItem, slotKey string) map[string]string {
	parsedResp := map[string]string{}
	if respItem == nil {
		return parsedResp
	}

	currentSlotKey := slotKey + "." + respItem.Key
	if strings.HasSuffix(slotKey, "-") {
		currentSlotKey = slotKey + respItem.Key
	}

	if len(respItem.Items) == 0 {
		parsedResp[currentSlotKey] = respItem.Value
		return parsedResp
	}

	for _, subItem := range respItem.Items {
		r := parseSlots(subItem, currentSlotKey)
		for k, v := range r {
			parsedResp[k] = v
		}
	}
	return parsedResp
}

func cleanUpConfidentialResponsesExports() {
	// clean up all existing confidential responses exports
	if err := filepath.Walk(conf.ExportPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Parse date from filename (assuming format YYYY-MM-DD##responses##..##..)
		basename := filepath.Base(path)
		parts := strings.Split(basename, "##")
		if len(parts) < 2 {
			return nil
		}

		if parts[1] != "confidential-responses" {
			return nil
		}

		// remove file
		err = os.Remove(path)
		if err != nil {
			slog.Error("Failed to remove old confidential responses export", slog.String("path", path), slog.String("error", err.Error()))
		}

		return nil
	}); err != nil {
		slog.Error("Error cleaning up old confidential responses exports", slog.String("error", err.Error()))
	}
}
