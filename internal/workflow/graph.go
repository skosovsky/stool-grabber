package workflow

import (
	"context"
	"fmt"

	"stool-grabber/internal/domain"
	"stool-grabber/internal/report"

	"github.com/skosovsky/flowy"
)

const (
	nodeScrape     = "scrape"
	nodeAggregate  = "aggregate"
	nodeAnalyze    = "analyze"
	nodeReportSkip = "report_skip"
	nodeReportLLM  = "report_llm"
	nodeReportAnalyzeError = "report_analyze_error"
	nodeFinish     = "finish"
)

func nextAfterAggregate(_ context.Context, state State) (string, error) {
	if state.Agg == nil {
		return "", fmt.Errorf("choice: aggregate result is nil")
	}
	if state.Agg.UsersAfterTopN == 0 {
		return nodeReportSkip, nil
	}
	return nodeAnalyze, nil
}

func nextAfterAnalyze(_ context.Context, state State) (string, error) {
	if state.Analyze != nil {
		return nodeReportLLM, nil
	}
	if state.AnalyzeErr != "" {
		return nodeReportAnalyzeError, nil
	}
	return "", fmt.Errorf("choice: analyze has neither result nor error")
}

// NewGraph builds and compiles the flowy graph.
func NewGraph(deps Deps) (*flowy.Graph[State], error) {
	if deps.Scraper == nil || deps.Aggregator == nil || deps.Analyzer == nil {
		return nil, fmt.Errorf("workflow deps must be non-nil")
	}

	b := flowy.NewGraph[State](Reducer)

	b.AddNode(nodeScrape, func(ctx context.Context, _ State) (State, error) {
		scrape, err := deps.Scraper.Scrape(ctx)
		if err != nil {
			return State{}, err
		}
		return State{Scrape: scrape}, nil
	})

	b.AddNode(nodeAggregate, func(ctx context.Context, state State) (State, error) {
		if state.Scrape == nil {
			return State{}, fmt.Errorf("aggregate: missing scrape")
		}
		agg, err := deps.Aggregator.Aggregate(ctx, state.Scrape)
		if err != nil {
			return State{}, err
		}
		return State{Agg: agg}, nil
	})

	b.AddNode(nodeAnalyze, func(ctx context.Context, state State) (State, error) {
		if state.Agg == nil {
			return State{}, fmt.Errorf("analyze: missing agg")
		}
		out, err := deps.Analyzer.Analyze(ctx, state.Agg)
		if err != nil {
			return State{AnalyzeErr: err.Error()}, nil
		}
		return State{Analyze: out}, nil
	})

	b.AddNode(nodeReportSkip, func(_ context.Context, _ State) (State, error) {
		return State{ReportMarkdown: report.RenderSkipReport()}, nil
	})

	b.AddNode(nodeReportLLM, func(_ context.Context, state State) (State, error) {
		if state.Analyze == nil {
			return State{}, fmt.Errorf("report_llm: missing analyze result")
		}
		users := map[int64]domain.UserRef(nil)
		if state.Scrape != nil {
			users = state.Scrape.Users
		}
		return State{ReportMarkdown: report.RenderLLMReport(state.ReportParams, state.Analyze, users)}, nil
	})

	b.AddNode(nodeReportAnalyzeError, func(_ context.Context, state State) (State, error) {
		return State{ReportMarkdown: report.RenderAnalyzeErrorReport(state.AnalyzeErr)}, nil
	})

	b.AddNode(nodeFinish, func(_ context.Context, _ State) (State, error) { return State{}, nil })

	b.AddEdge(nodeScrape, nodeAggregate)
	b.AddChoice(nodeAggregate, nextAfterAggregate)
	b.AddChoice(nodeAnalyze, nextAfterAnalyze)
	b.AddEdge(nodeReportSkip, nodeFinish)
	b.AddEdge(nodeReportLLM, nodeFinish)
	b.AddEdge(nodeReportAnalyzeError, nodeFinish)

	b.SetEntryPoint(nodeScrape)
	b.SetFinishPoint(nodeFinish)

	return b.Compile()
}

