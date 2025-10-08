package permissionchecker

const (
	SUBJECT_TYPE_MANAGEMENT_USER = "management-user"
	SUBJECT_TYPE_SERVICE_ACCOUNT = "service-account"
)

const (
	RESOURCE_TYPE_USERS     = "users"
	RESOURCE_TYPE_STUDY     = "study"
	RESOURCE_TYPE_MESSAGING = "messaging"
)

const (
	RESOURCE_KEY_STUDY_ALL = "*"

	RESOURCE_KEY_MESSAGING_GLOBAL_EMAIL_TEMPLATES = "global-email-templates"
	RESOURCE_KEY_MESSAGING_STUDY_EMAIL_TEMPLATES  = "study-email-templates"
	RESOURCE_KEY_MESSAGING_SCHEDULED_EMAILS       = "scheduled-emails"
	RESOURCE_KEY_MESSAGING_SMS_TEMPLATES          = "sms-templates"
)

const (
	ACTION_CREATE_STUDY                      = "create-study"
	ACTION_READ_STUDY_CONFIG                 = "read-study-config"
	ACTION_UPDATE_STUDY_PROPS                = "update-study-props"
	ACTION_UPDATE_STUDY_STATUS               = "update-study-status"
	ACTION_UPDATE_NOTIFICATION_SUBSCRIPTIONS = "update-notification-subscriptions"
	ACTION_UPDATE_STUDY_RULES                = "update-study-rules"
	ACTION_RUN_STUDY_ACTION                  = "run-study-action"
	ACTION_DELETE_STUDY                      = "delete-study"
	ACTION_MANAGE_STUDY_CODE_LISTS           = "manage-study-code-lists"
	ACTION_MANAGE_STUDY_COUNTERS             = "manage-study-counters"

	ACTION_MANAGE_STUDY_PERMISSIONS = "manage-study-permissions"

	ACTION_CREATE_SURVEY         = "create-survey"
	ACTION_UPDATE_SURVEY         = "update-survey"
	ACTION_UNPUBLISH_SURVEY      = "unpublish-survey"
	ACTION_DELETE_SURVEY_VERSION = "delete-survey-version"

	ACTION_GET_RESPONSES              = "get-responses"
	ACTION_DELETE_RESPONSES           = "delete-responses"
	ACTION_GET_CONFIDENTIAL_RESPONSES = "get-confidential-responses"
	ACTION_GET_FILES                  = "get-files"
	ACTION_DELETE_FILES               = "delete-files"
	ACTION_CREATE_VIRTUAL_PARTICIPANT = "create-virtual-participant"
	ACTION_EDIT_PARTICIPANT_DATA      = "edit-participant-data"
	ACTION_GET_PARTICIPANT_STATES     = "get-participant-states"
	ACTION_MERGE_PARTICIPANTS         = "merge-participants"
	ACTION_GET_REPORTS                = "get-reports"
	ACTION_UPDATE_REPORTS             = "update-reports"
	ACTION_DELETE_REPORTS             = "delete-reports"

	ACTION_DELETE_USERS = "delete-users"

	ACTION_ALL = "*"
)
