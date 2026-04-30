// Package cli implements CLI runner/presenter for the application.
package cli

// JobConfig is a narrow, CLI-oriented config derived from config.Job.
// It contains only what the runner needs.
type JobConfig struct {
	ChannelUsername string
	ParseDepth      int
	DelayMS         int

	MinCommentsToAnalyze int
	MinUniquePosts       int
	MaxUsersToAnalyze    int
	ExcludeAdmins        bool

	Model          string
	Temperature    float64
	TimeoutSeconds int
	PromptTemplate string

	OutputFilepath string
}

