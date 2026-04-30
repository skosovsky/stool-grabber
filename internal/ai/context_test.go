package ai

import (
	"context"
	"testing"

	"stool-grabber/internal/aggregate"
)

func TestBuildExecution_SetsModelAndTemperature(t *testing.T) {
	t.Parallel()

	params := AnalyzeParams{
		ChannelUsername: "@chan",
		Model:          "openrouter/test-model",
		Temperature:    0.2,
		PromptTemplate: "analyze",
	}

	agg := &aggregate.Result{
		Users: []aggregate.UserMessages{
			{UserID: "1", Username: "@u1", Messages: []string{"hello"}},
		},
	}

	prompts, err := NewEmbeddedPrompts()
	if err != nil {
		t.Fatalf("NewEmbeddedPrompts: %v", err)
	}

	exec, err := BuildExecution(context.Background(), prompts, params, agg)
	if err != nil {
		t.Fatalf("BuildExecution: %v", err)
	}
	if exec.ModelOptions == nil || exec.ModelOptions.Model != params.Model {
		t.Fatalf("model=%v, want %v", exec.ModelOptions, params.Model)
	}
	if exec.ModelOptions.Temperature == nil || *exec.ModelOptions.Temperature != params.Temperature {
		t.Fatalf("temperature=%v, want %v", exec.ModelOptions.Temperature, params.Temperature)
	}
	if exec.ResponseFormat == nil {
		t.Fatalf("ResponseFormat is nil; expected from manifest")
	}
}

