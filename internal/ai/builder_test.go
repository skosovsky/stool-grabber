package ai

import (
	"encoding/json"
	"testing"

	"stool-grabber/internal/aggregate"
)

func TestBuildAnalyzeCoreInput(t *testing.T) {
	t.Parallel()

	params := AnalyzeParams{
		ChannelUsername: "@test_chan",
		PromptTemplate:  "do the thing",
	}
	agg := &aggregate.Result{
		Users: []aggregate.UserMessages{
			{UserID: "10", Username: "@alice", Messages: []string{"a", "b"}},
			{UserID: "20", Username: "Bob Builder", Messages: []string{"c"}},
		},
	}

	id, in, err := BuildAnalyzeCoreInput(params, agg)
	if err != nil {
		t.Fatalf("BuildAnalyzeCoreInput() error: %v", err)
	}
	if id == "" {
		t.Fatalf("prompt id is empty")
	}
	if in.ChannelUsername != "test_chan" {
		t.Fatalf("ChannelUsername=%q, want %q", in.ChannelUsername, "test_chan")
	}
	if in.PromptTemplate != "do the thing" {
		t.Fatalf("PromptTemplate=%q, want %q", in.PromptTemplate, "do the thing")
	}
	if len(in.Users) != 2 {
		t.Fatalf("len(Users)=%d, want 2", len(in.Users))
	}
	if in.UsersJson == "" {
		t.Fatalf("UsersJson is empty")
	}

	var decoded []map[string]any
	if err := json.Unmarshal([]byte(in.UsersJson), &decoded); err != nil {
		t.Fatalf("UsersJson invalid JSON: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("decoded len=%d, want 2", len(decoded))
	}
}

