package cli

import "stool-grabber/internal/config"

func FromJob(job *config.Job) *JobConfig {
	if job == nil {
		return nil
	}
	return &JobConfig{
		ChannelUsername: job.Target.ChannelUsername,
		ParseDepth:      job.Target.ParseDepth,
		DelayMS:         job.Target.DelayMS,

		MinCommentsToAnalyze: job.Filter.MinCommentsToAnalyze,
		MinUniquePosts:       job.Filter.MinUniquePosts,
		MaxUsersToAnalyze:    job.Filter.MaxUsersToAnalyze,
		ExcludeAdmins:        job.Filter.ExcludeAdmins,

		Model:          job.LLM.Model,
		Temperature:    job.LLM.Temperature,
		PromptTemplate: job.LLM.PromptTemplate,

		OutputFilepath: job.Output.Filepath,
	}
}

