package smtp_client

type HeaderOverrides struct {
	From      string   `json:"from"`
	Sender    string   `json:"sender"`
	ReplyTo   []string `json:"replyTo"`
	NoReplyTo bool     `json:"noReplyTo"`
}
