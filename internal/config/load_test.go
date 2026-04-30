package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantErr   string
		assertJob func(t *testing.T, job *Job)
	}{
		{
			name: "valid config",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				path := filepath.Join(dir, "job.yaml")
				content := `version: "1.0"
target:
  channel_username: "yandex_go_dev"
  parse_depth: 100
  delay_ms: 1500
filter:
  min_comments_to_analyze: 4
  max_users_to_analyze: 20
  exclude_admins: true
llm:
  provider: "openrouter"
  model: "qwen/qwen3-30b-a3b"
  temperature: 0.2
  prompt_template: "Analyze"
output:
  format: "markdown"
  filepath: "./reports/target_channel_report.md"
`
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatalf("write config: %v", err)
				}
				return path
			},
			wantErr: "",
			assertJob: func(t *testing.T, job *Job) {
				t.Helper()
				if job.Filter.MaxUsersToAnalyze != 20 {
					t.Fatalf("expected max_users_to_analyze=20, got %d", job.Filter.MaxUsersToAnalyze)
				}
			},
		},
		{
			name: "missing file",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "missing.yaml")
			},
			wantErr: "config file not found",
		},
		{
			name: "invalid yaml",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				path := filepath.Join(dir, "job.yaml")
				if err := os.WriteFile(path, []byte("target: [broken"), 0o600); err != nil {
					t.Fatalf("write config: %v", err)
				}
				return path
			},
			wantErr: "decode job config",
		},
		{
			name: "validation error",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				path := filepath.Join(dir, "job.yaml")
				content := `version: "1.0"
target:
  channel_username: ""
  parse_depth: 10
  delay_ms: 100
filter:
  min_comments_to_analyze: 4
  max_users_to_analyze: 20
  exclude_admins: true
llm:
  provider: "openrouter"
  model: "qwen/qwen3-30b-a3b"
  temperature: 0.2
  prompt_template: "Analyze"
output:
  format: "markdown"
  filepath: "./reports/x.md"
`
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatalf("write config: %v", err)
				}
				return path
			},
			wantErr: "target.channel_username",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := tc.setup(t)
			job, err := Load(context.Background(), path)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if job == nil {
					t.Fatal("expected non-nil job")
				}
				if tc.assertJob != nil {
					tc.assertJob(t, job)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}
