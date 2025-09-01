package studyengine

import (
	"log/slog"

	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	STUDY_EVENT_TYPE_ENTER  = "ENTER"
	STUDY_EVENT_TYPE_SUBMIT = "SUBMIT"
	STUDY_EVENT_TYPE_TIMER  = "TIMER"
	STUDY_EVENT_TYPE_CUSTOM = "CUSTOM"
	STUDY_EVENT_TYPE_MERGE  = "MERGE"
	STUDY_EVENT_TYPE_LEAVE  = "LEAVE"
)

type StudyEngine struct {
	studyDBService   StudyDBService
	externalServices []ExternalService
	messageSender    StudyMessageSender
}

var (
	CurrentStudyEngine *StudyEngine
)

func InitStudyEngine(dbService StudyDBService, externalServices []ExternalService) {
	CurrentStudyEngine = &StudyEngine{
		studyDBService:   dbService,
		externalServices: externalServices,
	}
}

// RegisterStudyMessageSender allows to set a message sender implementation
// that can be swapped for simulator/test mode.
func (se *StudyEngine) RegisterStudyMessageSender(sender StudyMessageSender) {
	if CurrentStudyEngine != nil {
		CurrentStudyEngine.messageSender = sender
	} else {
		slog.Error("StudyEngine not initialized, cannot register message sender")
	}
}

type StudyDBService interface {
	GetResponses(instanceID string, studyKey string, filter bson.M, sort bson.M, page int64, limit int64) (responses []studyTypes.SurveyResponse, paginationInfo *studyDB.PaginationInfos, err error)
	DeleteConfidentialResponses(instanceID string, studyKey string, participantID string, key string) (count int64, err error)
	SaveResearcherMessage(instanceID string, studyKey string, message studyTypes.StudyMessage) error
	// Study code lists:
	StudyCodeListEntryExists(instanceID string, studyKey string, listKey string, code string) (bool, error)
	DeleteStudyCodeListEntry(instanceID string, studyKey string, listKey string, code string) error
	DrawStudyCode(instanceID string, studyKey string, listKey string) (string, error)
	// Study counters:
	GetCurrentStudyCounterValue(instanceID string, studyKey string, scope string) (int64, error)
	IncrementAndGetStudyCounterValue(instanceID string, studyKey string, scope string) (int64, error)
	RemoveStudyCounterValue(instanceID string, studyKey string, scope string) error
}

type ActionData struct {
	PState          studyTypes.Participant
	ReportsToCreate map[string]studyTypes.Report
}

type ExternalService struct {
	Name            string           `yaml:"name"`
	URL             string           `yaml:"url"`
	APIKey          string           `yaml:"apiKey"`
	Timeout         int              `yaml:"timeout"`
	MutualTLSConfig *MutualTLSConfig `yaml:"mTLSConfig"`
}

type MutualTLSConfig struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	CAFile   string `yaml:"caFile"`
}

type StudyEvent struct {
	InstanceID                            string
	StudyKey                              string
	Type                                  string                    // what kind of event (TIMER, SUBMISSION, ENTER etc.)
	Response                              studyTypes.SurveyResponse // if something is submitted during the event is added here
	Payload                               map[string]interface{}    // additional data
	EventKey                              string                    // key of the event	(for custom events)
	MergeWithParticipant                  studyTypes.Participant    // if need to merge with other participant state, is added here
	ParticipantIDForConfidentialResponses string
}

// EvalContext contains all the data that can be looked up by expressions
type EvalContext struct {
	Event            StudyEvent
	ParticipantState studyTypes.Participant
}

// SendOptions defines optional parameters for sending study emails.
type SendOptions struct {
	ExpiresAt        int64 // if message could not sent until this time, it will be discarded
	LanguageOverride string
}

// StudyMessageSender abstracts immediate message sending from the study engine.
// Implementations may send via SMTP bridge or capture messages in tests/simulators.
type StudyMessageSender interface {
	SendInstantStudyEmail(
		instanceID string,
		studyKey string,
		confidentialPID string,
		messageType string,
		extraPayload map[string]string,
		opts SendOptions,
	) error
}
