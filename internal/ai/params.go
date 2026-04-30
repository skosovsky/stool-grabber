package ai

// AnalyzeParams is a narrow DTO for AI analyze use-case.
type AnalyzeParams struct {
	ChannelUsername string
	Model          string
	Temperature    float64
	PromptTemplate string
}

