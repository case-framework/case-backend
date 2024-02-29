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
	ACTION_CREATE_STUDY = "create-study"

	ACTION_ALL = "*"
)
