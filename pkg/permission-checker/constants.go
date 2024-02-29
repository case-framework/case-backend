package permissionchecker

const (
	SUBJECT_TYPE_MANAGEMENT_USER = "management-user"
	SUBJECT_TYPE_SERVICE_ACCOUNT = "service-account"
)

const (
	RESOURCE_TYPE_STUDY     = "study"
	RESOURCE_TYPE_MESSAGING = "messaging"
)

const (
	RESOURCE_KEY_STUDY_ALL = "*"

	RESOURCE_KEY_MESSAGING_GLOBAL_EMAIL_TEMPLATES = "global-email-templates"
	RESOURCE_KEY_MESSAGING_STUDY_EMAIL_TEMPLATES  = "study-email-templates"
	RESOURCE_KEY_MESSAGING_SCHEDULED_EMAILS       = "scheduled-emails"
)

const (
	ACTION_CREATE_STUDY        = "create-study"
	ACTION_READ_STUDY_CONFIG   = "read-study-config"
	ACTION_UPDATE_STUDY_PROPS  = "update-study-props"
	ACTION_UPDATE_STUDY_STATUS = "update-study-status"
	ACTION_DELETE_STUDY        = "delete-study"

	ACTION_CREATE_SURVEY         = "create-survey"
	ACTION_UPDATE_SURVEY         = "update-survey"
	ACTION_UNPUBLISH_SURVEY      = "unpublish-survey"
	ACTION_DELETE_SURVEY_VERSION = "delete-survey-version"

	ACTION_ALL = "*"
)
