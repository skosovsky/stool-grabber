package workflow

import (
	"context"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"
	"stool-grabber/internal/domain"
)

type Scraper interface {
	Scrape(ctx context.Context) (*domain.ScrapeResult, error)
}

type Aggregator interface {
	Aggregate(ctx context.Context, scrape *domain.ScrapeResult) (*aggregate.Result, error)
}

type Analyzer interface {
	Analyze(ctx context.Context, agg *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error)
}

type ScraperFunc func(ctx context.Context) (*domain.ScrapeResult, error)

func (f ScraperFunc) Scrape(ctx context.Context) (*domain.ScrapeResult, error) { return f(ctx) }

type AggregatorFunc func(ctx context.Context, scrape *domain.ScrapeResult) (*aggregate.Result, error)

func (f AggregatorFunc) Aggregate(ctx context.Context, scrape *domain.ScrapeResult) (*aggregate.Result, error) {
	return f(ctx, scrape)
}

type AnalyzerFunc func(ctx context.Context, agg *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error)

func (f AnalyzerFunc) Analyze(ctx context.Context, agg *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error) {
	return f(ctx, agg)
}

type Deps struct {
	Scraper    Scraper
	Aggregator Aggregator
	Analyzer   Analyzer
}

