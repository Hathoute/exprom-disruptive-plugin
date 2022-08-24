package plugin

type queryModel struct {
	Entity        string            `json:"entity"`
	Parameters    map[string]string `json:"parameters"`
	WithStreaming bool              `json:"withStreaming"`
}
