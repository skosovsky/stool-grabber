package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai"
	"stool-grabber/internal/ai/contractgen"
	"stool-grabber/internal/domain"
	"stool-grabber/internal/report"
	"stool-grabber/internal/reportfs"
	"stool-grabber/internal/telegram"
	"stool-grabber/internal/workflow"

	"github.com/gotd/td/tg"
)

type Deps struct {
	TelegramCredentials telegram.Credentials

	Analyzer *ai.Analyzer

	In  io.Reader
	Out io.Writer
}

// RunJob runs the whole CLI job: auth, scrape, aggregate, optional analyze, render report,
// print stage summaries + report, and persist report file.
func RunJob(ctx context.Context, deps Deps, job *JobConfig) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	if deps.In == nil || deps.Out == nil {
		return fmt.Errorf("cli streams must be non-nil")
	}
	if deps.Analyzer == nil {
		return fmt.Errorf("analyzer is nil")
	}

	scrapeParams := telegram.ScrapeParams{
		ChannelUsername: job.ChannelUsername,
		ParseDepth:      job.ParseDepth,
		DelayMS:         job.DelayMS,
		ExcludeAdmins:   job.ExcludeAdmins,
	}
	aggregateParams := aggregate.Params{
		MinCommentsToAnalyze: job.MinCommentsToAnalyze,
		MaxUsersToAnalyze:    job.MaxUsersToAnalyze,
		ExcludeAdmins:        job.ExcludeAdmins,
	}
	analyzeParams := ai.AnalyzeParams{
		ChannelUsername: job.ChannelUsername,
		Model:          job.Model,
		Temperature:    job.Temperature,
		PromptTemplate: job.PromptTemplate,
	}
	reportParams := report.Params{
		ChannelUsername: job.ChannelUsername,
		Model:           job.Model,
	}

	return telegram.AuthorizedSessionRun(ctx, deps.TelegramCredentials, deps.In, deps.Out, func(runCtx context.Context, api *tg.Client) error {
		if err := telegram.PrintAuthorizedUser(runCtx, api, deps.Out); err != nil {
			return err
		}

		g, err := workflow.NewGraph(workflow.Deps{
			Scraper: workflow.ScraperFunc(func(c context.Context) (*domain.ScrapeResult, error) {
				return telegram.ScrapeChannelComments(c, api, scrapeParams)
			}),
			Aggregator: workflow.AggregatorFunc(func(_ context.Context, scrape *domain.ScrapeResult) (*aggregate.Result, error) {
				return aggregate.Aggregate(aggregateParams, scrape)
			}),
			Analyzer: workflow.AnalyzerFunc(func(c context.Context, agg *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error) {
				return deps.Analyzer.Analyze(c, analyzeParams, agg)
			}),
		})
		if err != nil {
			return err
		}

		final, err := g.Invoke(runCtx, workflow.State{ReportParams: reportParams})
		if err != nil {
			return err
		}

		if final.Scrape != nil {
			_, _ = fmt.Fprintf(deps.Out, "Scrape summary: posts=%d, comments=%d (discussion_chat_id=%d, channel=@%s).\n",
				len(final.Scrape.Threads),
				countComments(final.Scrape),
				final.Scrape.LinkedDiscussionChatID,
				final.Scrape.ChannelUsername,
			)
		}
		if final.Agg != nil {
			_, _ = fmt.Fprintf(deps.Out, "Aggregate summary: users_before_topN=%d, users_after_topN=%d, json_bytes=%d.\n",
				final.Agg.UsersBeforeTopN,
				final.Agg.UsersAfterTopN,
				len(final.Agg.UsersJSON),
			)
		}
		if final.Analyze != nil {
			_, _ = fmt.Fprintf(deps.Out, "Analyze summary: agitators=%d, hot_topics=%d.\n", len(final.Analyze.Agitators), len(final.Analyze.HotTopics))
		} else {
			_, _ = fmt.Fprintln(deps.Out, "Analyze skipped: no users after filtering/top-N.")
		}

		if final.ReportMarkdown != "" {
			_, _ = fmt.Fprintln(deps.Out, final.ReportMarkdown)
		}
		if job.OutputFilepath != "" && final.ReportMarkdown != "" {
			if err := reportfs.WriteMarkdownFile(job.OutputFilepath, final.ReportMarkdown); err != nil {
				return err
			}
		}

		if job.OutputFilepath != "" && final.Scrape != nil {
			td, _, err := report.BuildTechnicalDump(final.Scrape)
			if err != nil {
				return err
			}
			if err := reportfs.WriteJSONFile(techDumpPath(job.OutputFilepath), td); err != nil {
				return err
			}
		}
		return nil
	})
}

func countComments(res *domain.ScrapeResult) int {
	if res == nil {
		return 0
	}
	n := 0
	for _, t := range res.Threads {
		n += len(t.Comments)
	}
	return n
}

func techDumpPath(markdownPath string) string {
	ext := strings.ToLower(filepath.Ext(markdownPath))
	base := strings.TrimSuffix(markdownPath, ext)
	if base == markdownPath {
		return markdownPath + ".tech.json"
	}
	return base + ".tech.json"
}

