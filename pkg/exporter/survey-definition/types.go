package surveydefinition

type ExtractOptions struct {
	IncludeItems []string
	ExcludeItems []string
	UseLabelLang string
}

type SurveyVersionPreview struct {
	VersionID   string           `json:"versionId"`
	Published   int64            `json:"published"`
	Unpublished int64            `json:"unpublished"`
	Questions   []SurveyQuestion `json:"questions"`
}

type SurveyQuestion struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	QuestionType string        `json:"questionType"`
	Responses    []ResponseDef `json:"responses"`
}

type ResponseDef struct {
	ID           string           `json:"id"`
	ResponseType string           `json:"responseType"`
	Label        string           `json:"label"`
	Options      []ResponseOption `json:"options"`
}

type ResponseOption struct {
	ID         string `json:"id"`
	OptionType string `json:"optionType"`
	Label      string `json:"label"`
}
