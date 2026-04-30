// Package workflow defines the pure orchestration graph (flowy) and its state.
package workflow

import (
	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"
	"stool-grabber/internal/domain"
	"stool-grabber/internal/report"
)

// State is the workflow state carried through flowy graph nodes.
// Use delta updates: nodes return State with only modified fields.
type State struct {
	Scrape  *domain.ScrapeResult
	Agg     *aggregate.Result
	Analyze *contractgen.AnalyzeCoreOutput
	AnalyzeErr string

	ReportParams   report.Params
	ReportMarkdown string
}

// Reducer merges node updates into current state.
func Reducer(current, update State) State {
	if update.Scrape != nil {
		current.Scrape = update.Scrape
	}
	if update.Agg != nil {
		current.Agg = update.Agg
	}
	if update.Analyze != nil {
		current.Analyze = update.Analyze
	}
	if update.AnalyzeErr != "" {
		current.AnalyzeErr = update.AnalyzeErr
	}
	if update.ReportParams != (report.Params{}) {
		current.ReportParams = update.ReportParams
	}
	if update.ReportMarkdown != "" {
		current.ReportMarkdown = update.ReportMarkdown
	}
	return current
}

