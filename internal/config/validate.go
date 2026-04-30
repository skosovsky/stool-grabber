package config

import (
	"fmt"
	"strings"
)

func Validate(job Job) error {
	if job.Target.ParseDepth <= 0 {
		return validationError("target.parse_depth", "must be > 0")
	}

	if job.Target.DelayMS < 0 {
		return validationError("target.delay_ms", "must be >= 0")
	}

	if job.Filter.MinCommentsToAnalyze < 1 {
		return validationError("filter.min_comments_to_analyze", "must be >= 1")
	}

	if job.Filter.MaxUsersToAnalyze < 1 {
		return validationError("filter.max_users_to_analyze", "must be >= 1")
	}

	if strings.TrimSpace(job.Target.ChannelUsername) == "" {
		return validationError("target.channel_username", "must not be empty")
	}

	if strings.TrimSpace(job.LLM.Model) == "" {
		return validationError("llm.model", "must not be empty")
	}

	if strings.TrimSpace(job.Output.Filepath) == "" {
		return validationError("output.filepath", "must not be empty")
	}

	return nil
}

func validationError(field, rule string) error {
	return fmt.Errorf("invalid config: %s %s", field, rule)
}
