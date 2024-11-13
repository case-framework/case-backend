package apihandlers

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"

	"github.com/gin-gonic/gin"
)

const (
	EML_FILE_EXTENSION = ".eml"
	EMAILS_DIR         = "emails"
)

type SendEmailReq struct {
	To              []string                        `json:"to"`
	Subject         string                          `json:"subject"`
	Content         string                          `json:"content"`
	HighPrio        bool                            `json:"highPrio"`
	HeaderOverrides *messagingTypes.HeaderOverrides `json:"headerOverrides"`
}

func (h *HttpEndpoints) AddRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/")

	auth.POST("/send-email",
		mw.HasValidAPIKey(h.apiKeys),
		h.sendEmail)
}

// creates a directory to store emails (.eml) and handles errors.
func createEmailsDirectory() error {
	if err := os.MkdirAll(EMAILS_DIR, os.ModePerm); err != nil {
		slog.Error("Error creating directory: "+EMAILS_DIR, slog.String("error", err.Error()))
		return err
	}
	return nil
}

// saves an email to a .eml file
func saveEmailAsEML(email SendEmailReq) error {
	if err := createEmailsDirectory(); err != nil {
		return err
	}

	// Generate a unique EML file path
	emlFilePath, err := getUniqueEMLFilePath()
	if err != nil {
		slog.Error("Error generating EML file path", slog.String("FilePath", emlFilePath), slog.String("error", err.Error()))
		return err
	}

	// Create .eml file
	file, err := os.Create(emlFilePath)
	if err != nil {
		slog.Error("Failed to create EML file ", slog.String("path", emlFilePath), slog.String("error", err.Error()))
		return err
	}
	defer file.Close()

	//compose email headers and write to file
	headers := composeHeaders(email)
	if err := writeHeadersToFile(file, headers); err != nil {
		slog.Error("Failed to write headers to file", slog.String("FilePath", emlFilePath), slog.String("error", err.Error()))
		return err
	}

	// Write content (body) to file
	_, err = file.WriteString(email.Content)
	if err != nil {
		slog.Error("Failed to write email content", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Email has been saved as EML file successfully")
	return nil
}

// constructs the email headers as a map from SendEmailReq
func composeHeaders(email SendEmailReq) map[string]string {
	headers := map[string]string{
		"To":      strings.Join(email.To, ", "),
		"Subject": email.Subject,
		"Date":    time.Now().Format(time.RFC1123Z),
	}

	// Priority header
	if email.HighPrio {
		headers["Priority"] = "High"
	}

	// Optional header overrides
	if email.HeaderOverrides != nil {
		if email.HeaderOverrides.From != "" {
			headers["From"] = email.HeaderOverrides.From
		}
		if email.HeaderOverrides.Sender != "" {
			headers["Sender"] = email.HeaderOverrides.Sender
		}
		if !email.HeaderOverrides.NoReplyTo && len(email.HeaderOverrides.ReplyTo) > 0 {
			headers["Reply-To"] = strings.Join(email.HeaderOverrides.ReplyTo, ", ")
		}
	}

	return headers
}

// writes headers map to the provided file
func writeHeadersToFile(file *os.File, headers map[string]string) error {
	for key, value := range headers {
		if _, err := file.WriteString(key + ": " + value + "\n"); err != nil {
			slog.Error("failed to write header", slog.String("key", key), slog.String("error", err.Error()))
			return err
		}
	}
	// Separate headers from body
	if _, err := file.WriteString("\n"); err != nil {
		slog.Error("Failed to write header-body separator", slog.String("error", err.Error()))
		return err
	}
	return nil
}

// returns a unique file path for the EML file i.e. timestamp.eml, but appending a counter if needed.
func getUniqueEMLFilePath() (string, error) {
	baseFileName := getEMLFilename()
	emlFilePath := filepath.Join(EMAILS_DIR, baseFileName)
	counter := 1

	// Check if the file already exists, and append a counter if necessary
	for {
		if _, err := os.Stat(emlFilePath); errors.Is(err, os.ErrNotExist) {
			break
		}
		baseName := filepath.Base(emlFilePath)
		ext := filepath.Ext(emlFilePath)
		baseNameWithoutExt := baseName[:len(baseName)-len(ext)]

		emlFilePath = filepath.Join(EMAILS_DIR, baseNameWithoutExt+"_"+strconv.Itoa(counter)+EML_FILE_EXTENSION)
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
		slog.Error("Email could not be saved as EML file", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Email could not be saved as EML file"})
		return
	}

	slog.Info("Email has been saved as EML file")
	c.JSON(http.StatusOK, gin.H{"message": "Email has been saved as EML file"})
}
