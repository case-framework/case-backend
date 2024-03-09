package surveyresponses

type ParsedResponse struct {
	ID            string
	ParticipantID string
	OpenedAt      int64
	SubmittedAt   int64
	Version       string
	Context       map[string]string // e.g. Language, or engine version
	Responses     map[string]interface{}
	Meta          ResponseMeta
}

type ResponseMeta struct {
	Initialised map[string][]int64
	Displayed   map[string][]int64
	Responded   map[string][]int64
	Position    map[string]int32
}

type IncludeMeta struct {
	Postion        bool
	InitTimes      bool
	DisplayedTimes bool
	ResponsedTimes bool
}

type ColumnNames struct {
	FixedColumns    []string
	ContextColumns  []string
	ResponseColumns []string
	MetaColumns     []string
}
