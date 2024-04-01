package surveydefinition

const (
	SURVEY_ITEM_COMPONENT_ROLE_TITLE          = "title"
	SURVEY_ITEM_COMPONENT_ROLE_RESPONSE_GROUP = "responseGroup"
)

const (
	QUESTION_TYPE_CONSENT                         = "consent"
	QUESTION_TYPE_SINGLE_CHOICE                   = "single_choice"
	QUESTION_TYPE_MULTIPLE_CHOICE                 = "multiple_choice"
	QUESTION_TYPE_TEXT_INPUT                      = "text"
	QUESTION_TYPE_NUMBER_INPUT                    = "number"
	QUESTION_TYPE_DATE_INPUT                      = "date"
	QUESTION_TYPE_DROPDOWN                        = "dropdown"
	QUESTION_TYPE_LIKERT                          = "likert"
	QUESTION_TYPE_LIKERT_GROUP                    = "likert_group"
	QUESTION_TYPE_EQ5D_SLIDER                     = "eq5d_slider"
	QUESTION_TYPE_NUMERIC_SLIDER                  = "slider"
	QUESTION_TYPE_RESPONSIVE_TABLE                = "responsive_table"
	QUESTION_TYPE_MATRIX                          = "matrix"
	QUESTION_TYPE_MATRIX_RADIO_ROW                = "matrix_radio_row"
	QUESTION_TYPE_MATRIX_DROPDOWN                 = "matrix_dropdown"
	QUESTION_TYPE_MATRIX_INPUT                    = "matrix_input"
	QUESTION_TYPE_MATRIX_NUMBER_INPUT             = "matrix_number_input"
	QUESTION_TYPE_MATRIX_CHECKBOX                 = "matrix_checkbox"
	QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY  = "responsive_single_choice_array"
	QUESTION_TYPE_RESPONSIVE_BIPOLAR_LIKERT_ARRAY = "responsive_bipolar_likert_array"
	QUESTION_TYPE_CLOZE                           = "cloze"
	QUESTION_TYPE_UNKNOWN                         = "unknown"
	QUESTION_TYPE_EMPTY                           = "empty"
)

const (
	OPTION_TYPE_DROPDOWN_OPTION             = "option"
	OPTION_TYPE_RADIO                       = "radio"
	OPTION_TYPE_CHECKBOX                    = "checkbox"
	OPTION_TYPE_TEXT_INPUT                  = "text"
	OPTION_TYPE_DATE_INPUT                  = "date"
	OPTION_TYPE_NUMBER_INPUT                = "number"
	OPTION_TYPE_CLOZE                       = "cloze"
	OPTION_TYPE_DROPDOWN                    = "dropdown"
	OPTION_TYPE_EMBEDDED_CLOZE_TEXT_INPUT   = "embedded_cloze_text"
	OPTION_TYPE_EMBEDDED_CLOZE_DATE_INPUT   = "embedded_cloze_date"
	OPTION_TYPE_EMBEDDED_CLOZE_NUMBER_INPUT = "embedded_cloze_number"
	OPTION_TYPE_EMBEDDED_CLOZE_DROPDOWN     = "embedded_cloze_dropdown"
)

const (
	RESPONSE_ROOT_KEY = "rg"
)

const (
	OPEN_FIELD_COL_SUFFIX = "open"
	TRUE_VALUE            = "TRUE"
	FALSE_VALUE           = "FALSE"
)
