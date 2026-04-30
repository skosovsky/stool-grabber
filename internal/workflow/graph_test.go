package workflow

import (
	"context"
	"sync/atomic"
	"testing"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"
	"stool-grabber/internal/domain"
)

func TestNextAfterAggregate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero_users", 0, nodeReportSkip},
		{"some_users", 1, nodeAnalyze},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			next, err := nextAfterAggregate(context.Background(), State{
				Agg: &aggregate.Result{UsersAfterTopN: tt.n},
			})
			if err != nil {
				t.Fatalf("nextAfterAggregate error: %v", err)
			}
			if next != tt.want {
				t.Fatalf("next=%q, want %q", next, tt.want)
			}
		})
	}
}

func TestGraph_DoesNotCallAnalyzeWhenEmpty(t *testing.T) {
	t.Parallel()

	var analyzeCalls atomic.Int64

	deps := Deps{
		Scraper: ScraperFunc(func(_ context.Context) (*domain.ScrapeResult, error) {
			return &domain.ScrapeResult{}, nil
		}),
		Aggregator: AggregatorFunc(func(_ context.Context, _ *domain.ScrapeResult) (*aggregate.Result, error) {
			return &aggregate.Result{UsersAfterTopN: 0}, nil
		}),
		Analyzer: AnalyzerFunc(func(_ context.Context, _ *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error) {
			analyzeCalls.Add(1)
			return &contractgen.AnalyzeCoreOutput{}, nil
		}),
	}

	g, err := NewGraph(deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	final, err := g.Invoke(context.Background(), State{})
	_ = final
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if got := analyzeCalls.Load(); got != 0 {
		t.Fatalf("analyzeCalls=%d, want 0", got)
	}
}

