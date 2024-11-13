package apihandlers

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"

	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/")

	auth.POST("/send-email",
		mw.HasValidAPIKey(h.apiKeys),
		h.sendEmail)
}

const (
	EML_FILE_EXTENSION = ".eml"
)

type SendEmailReq struct {
	To              []string                        `json:"to"`
	Subject         string                          `json:"subject"`
	Content         string                          `json:"content"`
	HighPrio        bool                            `json:"highPrio"`
	HeaderOverrides *messagingTypes.HeaderOverrides `json:"headerOverrides"`
}

// creates a directory for the recipient and handles errors.
func createFolder(folderPath, recipient string) error {
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		slog.Error("Error creating folder for recipient", slog.String("recipient", recipient), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func saveEmailAsEML(email SendEmailReq) error {
	for _, recipient := range email.To {

		// Create the recipient's folder and handle any error
		folderPath := filepath.Join("emails", recipient)

		if err := createFolder(folderPath, recipient); err != nil {
			return err
		}

		// Generate a unique EML file path
		emlFilePath, err := getUniqueEMLFilePath(folderPath)
		if err != nil {
			slog.Error("Error generating EML file path", slog.String("recipient", recipient), slog.String("error", err.Error()))
			return err
		}

		// Write the EML content to the file
		if err := os.WriteFile(emlFilePath, []byte(email.Content), 0644); err != nil {
			slog.Error("Error writing EML file for "+recipient, slog.String("error", err.Error()))
			return err
		}

		slog.Info("Successfully created folder '" + recipient + "' and its content as EML file inside the folder")
	}

	return nil
}

// returns a unique file path for the EML file, appending a counter if needed.
func getUniqueEMLFilePath(folderPath string) (string, error) {
	baseFileName := getEMLFilename()
	emlFilePath := filepath.Join(folderPath, baseFileName)
	counter := 1

	// Check if the file already exists, and append a counter if necessary
	for {
		if _, err := os.Stat(emlFilePath); errors.Is(err, os.ErrNotExist) {
			break
		}
		baseName := filepath.Base(emlFilePath)
		ext := filepath.Ext(emlFilePath)
		baseNameWithoutExt := baseName[:len(baseName)-len(ext)]

		emlFilePath = filepath.Join(folderPath, baseNameWithoutExt+"_"+strconv.Itoa(counter)+EML_FILE_EXTENSION)
		counter++
	}

	return emlFilePath, nil
}

// generates a valid file name for the EML file
func getEMLFilename() string {

	timestamp := time.Now().Format("20060102_150405")
	fileName := timestamp + EML_FILE_EXTENSION

	return fileName
}

func (h *HttpEndpoints) sendEmail(c *gin.Context) {
	var req SendEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.To) < 1 {
		slog.Error("missing 'to' field")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'to' field"})
		return
	}

	if err := saveEmailAsEML(req); err != nil {
		slog.Error("Email could not be saved as EML file(s)", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Email could not be saved as EML file(s)"})
		return
	}

	slog.Info("Email has been saved as EML file(s)")
	c.JSON(http.StatusOK, gin.H{"message": "Email has been saved as EML file(s)"})
}
