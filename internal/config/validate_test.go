package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	valid := Job{
		Version: "1.0",
		Target: JobTarget{
			ChannelUsername: "yandex_go_dev",
			ParseDepth:      100,
			DelayMS:         1500,
		},
		Filter: JobFilter{
			MinCommentsToAnalyze: 4,
			MaxUsersToAnalyze:    20,
			ExcludeAdmins:        true,
		},
		LLM: JobLLM{
			Provider:       "openrouter",
			Model:          "qwen/qwen3-30b-a3b",
			Temperature:    0.2,
			PromptTemplate: "Analyze messages",
		},
		Output: JobOutput{
			Format:   "markdown",
			Filepath: "./reports/report.md",
		},
	}

	tests := []struct {
		name    string
		mutate  func(j Job) Job
		wantErr string
	}{
		{
			name: "valid",
			mutate: func(j Job) Job {
				return j
			},
			wantErr: "",
		},
		{
			name: "invalid parse_depth",
			mutate: func(j Job) Job {
				j.Target.ParseDepth = 0
				return j
			},
			wantErr: "target.parse_depth",
		},
		{
			name: "invalid delay_ms",
			mutate: func(j Job) Job {
				j.Target.DelayMS = -1
				return j
			},
			wantErr: "target.delay_ms",
		},
		{
			name: "invalid min_comments_to_analyze",
			mutate: func(j Job) Job {
				j.Filter.MinCommentsToAnalyze = 0
				return j
			},
			wantErr: "filter.min_comments_to_analyze",
		},
		{
			name: "invalid max_users_to_analyze",
			mutate: func(j Job) Job {
				j.Filter.MaxUsersToAnalyze = 0
				return j
			},
			wantErr: "filter.max_users_to_analyze",
		},
		{
			name: "empty channel username",
			mutate: func(j Job) Job {
				j.Target.ChannelUsername = "   "
				return j
			},
			wantErr: "target.channel_username",
		},
		{
			name: "empty llm model",
			mutate: func(j Job) Job {
				j.LLM.Model = ""
				return j
			},
			wantErr: "llm.model",
		},
		{
			name: "empty output filepath",
			mutate: func(j Job) Job {
				j.Output.Filepath = ""
				return j
			},
			wantErr: "output.filepath",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			job := tc.mutate(valid)
			err := Validate(job)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if got := err.Error(); !strings.Contains(got, tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, got)
				}
			}
		})
	}
}
