package workflow

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"
	"stool-grabber/internal/domain"
	"stool-grabber/internal/report"
)

func TestWorkflow_FullPipeline(t *testing.T) {
	t.Parallel()

	var analyzeCalls atomic.Int64

	g, err := NewGraph(Deps{
		Scraper: ScraperFunc(func(_ context.Context) (*domain.ScrapeResult, error) {
			return &domain.ScrapeResult{
				ChannelUsername: "chan",
				Users: map[int64]domain.UserRef{
					10: {ID: 10, Username: "u10", FirstName: "A", LastName: "B"},
				},
				Threads: []domain.PostThread{
					{ChannelMessageID: 1, Comments: []domain.Comment{{SenderUserID: 10, Text: "x"}}},
				},
			}, nil
		}),
		Aggregator: AggregatorFunc(func(_ context.Context, _ *domain.ScrapeResult) (*aggregate.Result, error) {
			return &aggregate.Result{UsersAfterTopN: 1}, nil
		}),
		Analyzer: AnalyzerFunc(func(_ context.Context, _ *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error) {
			analyzeCalls.Add(1)
			return &contractgen.AnalyzeCoreOutput{
				Agitators: []contractgen.AnalyzeCoreOutputAgitatorsItem{
					{User: "10", MessageCount: 1, WhatTriggers: "a", Style: "b", WontPassBy: "c"},
				},
				HotTopics: []string{"t1"},
			}, nil
		}),
	})
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	final, err := g.Invoke(context.Background(), State{
		ReportParams: report.Params{ChannelUsername: "@chan", Model: "m"},
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if got := analyzeCalls.Load(); got != 1 {
		t.Fatalf("analyzeCalls=%d, want 1", got)
	}
	if final.ReportMarkdown == "" {
		t.Fatalf("ReportMarkdown is empty")
	}
	if !strings.Contains(final.ReportMarkdown, "# Отчёт") {
		t.Fatalf("missing report header: %q", final.ReportMarkdown)
	}
	if !strings.Contains(final.ReportMarkdown, "## Agitators") {
		t.Fatalf("missing agitators section: %q", final.ReportMarkdown)
	}
	if !strings.Contains(final.ReportMarkdown, "## Hot topics") {
		t.Fatalf("missing hot topics section: %q", final.ReportMarkdown)
	}
	if !strings.Contains(final.ReportMarkdown, "@u10") {
		t.Fatalf("missing username: %q", final.ReportMarkdown)
	}
	if !strings.Contains(final.ReportMarkdown, "A B") {
		t.Fatalf("missing full name: %q", final.ReportMarkdown)
	}
}

func TestWorkflow_PostsWithoutComments_SkipBranch(t *testing.T) {
	t.Parallel()

	var analyzeCalls atomic.Int64

	g, err := NewGraph(Deps{
		Scraper: ScraperFunc(func(_ context.Context) (*domain.ScrapeResult, error) {
			return &domain.ScrapeResult{
				ChannelUsername: "chan",
				Threads: []domain.PostThread{
					{ChannelMessageID: 1, Comments: nil},
					{ChannelMessageID: 2, Comments: []domain.Comment{}},
				},
			}, nil
		}),
		Aggregator: AggregatorFunc(func(_ context.Context, _ *domain.ScrapeResult) (*aggregate.Result, error) {
			return &aggregate.Result{UsersAfterTopN: 0}, nil
		}),
		Analyzer: AnalyzerFunc(func(_ context.Context, _ *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error) {
			analyzeCalls.Add(1)
			return &contractgen.AnalyzeCoreOutput{}, nil
		}),
	})
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	final, err := g.Invoke(context.Background(), State{})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if got := analyzeCalls.Load(); got != 0 {
		t.Fatalf("analyzeCalls=%d, want 0", got)
	}
	if final.ReportMarkdown == "" {
		t.Fatalf("ReportMarkdown is empty")
	}
	if !strings.Contains(final.ReportMarkdown, "Целевое ядро не выявлено") {
		t.Fatalf("unexpected skip report: %q", final.ReportMarkdown)
	}
}

func TestWorkflow_AllUsersFiltered_SkipBranch(t *testing.T) {
	t.Parallel()

	var analyzeCalls atomic.Int64

	g, err := NewGraph(Deps{
		Scraper: ScraperFunc(func(_ context.Context) (*domain.ScrapeResult, error) {
			return &domain.ScrapeResult{
				ChannelUsername: "chan",
				Threads: []domain.PostThread{
					{ChannelMessageID: 1, Comments: []domain.Comment{{SenderUserID: 10, Text: "x"}}},
				},
			}, nil
		}),
		Aggregator: AggregatorFunc(func(_ context.Context, _ *domain.ScrapeResult) (*aggregate.Result, error) {
			// Simulate filters/top-N removing everyone.
			return &aggregate.Result{UsersAfterTopN: 0}, nil
		}),
		Analyzer: AnalyzerFunc(func(_ context.Context, _ *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error) {
			analyzeCalls.Add(1)
			return &contractgen.AnalyzeCoreOutput{}, nil
		}),
	})
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	_, err = g.Invoke(context.Background(), State{})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if got := analyzeCalls.Load(); got != 0 {
		t.Fatalf("analyzeCalls=%d, want 0", got)
	}
}

