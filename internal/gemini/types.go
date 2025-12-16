package gemini

// TokenUsage contains token usage information from API response
type TokenUsage struct {
	Prompt     int `yaml:"prompt"`
	Candidates int `yaml:"candidates"`
	Total      int `yaml:"total"`
	Cached     int `yaml:"cached,omitempty"`
	Thoughts   int `yaml:"thoughts,omitempty"`
}
