package config

type Job struct {
	Version string       `yaml:"version"`
	Target  JobTarget    `yaml:"target"`
	Filter  JobFilter    `yaml:"filter"`
	LLM     JobLLM       `yaml:"llm"`
	Output  JobOutput    `yaml:"output"`
}

type JobTarget struct {
	ChannelUsername string `yaml:"channel_username"`
	ParseDepth      int    `yaml:"parse_depth"`
	DelayMS         int    `yaml:"delay_ms"`
}

type JobFilter struct {
	MinCommentsToAnalyze int  `yaml:"min_comments_to_analyze"`
	MinUniquePosts       int  `yaml:"min_unique_posts"`
	MaxUsersToAnalyze    int  `yaml:"max_users_to_analyze"`
	ExcludeAdmins        bool `yaml:"exclude_admins"`
}

type JobLLM struct {
	Provider       string  `yaml:"provider"`
	Model          string  `yaml:"model"`
	Temperature    float64 `yaml:"temperature"`
	TimeoutSeconds int     `yaml:"timeout_seconds"`
	PromptTemplate string  `yaml:"prompt_template"`
}

type JobOutput struct {
	Format   string `yaml:"format"`
	Filepath string `yaml:"filepath"`
}
