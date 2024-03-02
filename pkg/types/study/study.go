package study

const (
	STUDY_STATUS_ACTIVE   = "active"
	STUDY_STATUS_INACTIVE = "inactive"
)

const (
	DEFAULT_ID_MAPPING_METHOD = "sha-256"
)

type Study struct {
	ID                        string                     `bson:"_id" json:"id,omitempty"`
	Key                       string                     `bson:"key" json:"key"`
	SecretKey                 string                     `bson:"secretKey" json:"secretKey"`
	Status                    string                     `bson:"status" json:"status"`
	Props                     StudyProps                 `bson:"props" json:"props"`
	Configs                   StudyConfigs               `bson:"configs" json:"configs"`
	NotificationSubscriptions []NotificationSubscription `bson:"notificationSubscriptions" json:"notificationSubscriptions"`

	// depracted fields potentially to be removed in the future
	Stats          StudyStats   `bson:"studyStats" json:"stats"`
	NextTimerEvent int64        `bson:"nextTimerEvent" json:"nextTimerEvent"`
	Rules          []Expression `bson:"rules" json:"rules"`
}

type StudyProps struct {
	Name               []LocalisedObject `bson:"name" json:"name"`
	Description        []LocalisedObject `bson:"description" json:"description"`
	Tags               []Tag             `bson:"tags" json:"tags"`
	StartDate          int64             `bson:"startDate" json:"startDate"`
	EndDate            int64             `bson:"endDate" json:"endDate"`
	SystemDefaultStudy bool              `bson:"systemDefaultStudy" json:"systemDefaultStudy"`
}

type StudyConfigs struct {
	ParticipantFileUploadRule *Expression `bson:"participantFileUploadRule" json:"participantFileUploadRule"`
	IdMappingMethod           string      `bson:"idMappingMethod" json:"idMappingMethod"`
}

type StudyStats struct {
	ParticipantCount     int64 `bson:"participantCount" json:"participantCount"`
	TempParticipantCount int64 `bson:"tempParticipantCount" json:"tempParticipantCount"`
	ResponseCount        int64 `bson:"responseCount" json:"responseCount"`
}

type Tag struct {
	Label []LocalisedObject `bson:"label" json:"label"`
}

type NotificationSubscription struct {
	MessageType string `bson:"messageType" json:"messageType"`
	Email       string `bson:"email" json:"email"`
}
