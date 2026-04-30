package ai

import (
	"context"
	"iter"
	"testing"

	"stool-grabber/internal/ai/contractgen"

	"github.com/skosovsky/prompty"
)

func TestAnalyzeWithInvoker_ParsesStructuredOutput(t *testing.T) {
	t.Parallel()

	exec := prompty.SimpleChat(
		"You are a system.",
		"Return JSON.",
	)
	exec.ResponseFormat = &prompty.SchemaDefinition{
		Name: "dummy",
		Schema: map[string]any{
			"type": "object",
		},
	}

	inv := &fakeInvoker{
		text: `{"agitators":[{"user":"10","message_count":3,"what_triggers":"x","style":"y","wont_pass_by":"z"}],"hot_topics":["t1","t2"]}`,
	}

	got, err := analyzeWithInvoker(context.Background(), inv, exec)
	if err != nil {
		t.Fatalf("analyzeWithInvoker error: %v", err)
	}
	if got == nil {
		t.Fatalf("nil output")
	}
	if len(got.Agitators) != 1 {
		t.Fatalf("agitators=%d, want 1", len(got.Agitators))
	}
	if got.Agitators[0].User != "10" {
		t.Fatalf("user=%q, want %q", got.Agitators[0].User, "10")
	}
	if got.Agitators[0].MessageCount != 3 {
		t.Fatalf("message_count=%d, want 3", got.Agitators[0].MessageCount)
	}
	if len(got.HotTopics) != 2 {
		t.Fatalf("hot_topics=%d, want 2", len(got.HotTopics))
	}
}

type fakeInvoker struct {
	text string
}

func (f *fakeInvoker) Execute(_ context.Context, _ *prompty.PromptExecution) (*prompty.Response, error) {
	return &prompty.Response{
		Content: []prompty.ContentPart{
			prompty.TextPart{Text: f.text},
		},
	}, nil
}

func (f *fakeInvoker) ExecuteStream(_ context.Context, _ *prompty.PromptExecution) iter.Seq2[*prompty.ResponseChunk, error] {
	return func(yield func(*prompty.ResponseChunk, error) bool) {
		yield(&prompty.ResponseChunk{
			Content: []prompty.ContentPart{prompty.TextPart{Text: f.text}},
		}, nil)
		yield(&prompty.ResponseChunk{IsFinished: true}, nil)
	}
}

var _ prompty.Invoker = (*fakeInvoker)(nil)

// compile-time check: ensure output type stays from contractgen
var _ *contractgen.AnalyzeCoreOutput = nil

