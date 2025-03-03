package studyengine

import (
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

type StudyDBService interface {
	GetResponses(instanceID string, studyKey string, filter bson.M, sort bson.M, page int64, limit int64) (responses []studyTypes.SurveyResponse, paginationInfo *studyDB.PaginationInfos, err error)
	DeleteConfidentialResponses(instanceID string, studyKey string, participantID string, key string) (count int64, err error)
	SaveResearcherMessage(instanceID string, studyKey string, message studyTypes.StudyMessage) error
	StudyCodeListEntryExists(instanceID string, studyKey string, listKey string, code string) (bool, error)
	DeleteStudyCodeListEntry(instanceID string, studyKey string, listKey string, code string) error
	DrawStudyCode(instanceID string, studyKey string, listKey string) (string, error)
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
