package types

type RenderOptions struct {
	Query  string
	Params map[string]string
}

type SilenceCreateResponse struct {
	SilenceID string `json:"silenceID"`
}

type QueryMatcher struct {
	Key      string
	Operator string
	Value    string
}
