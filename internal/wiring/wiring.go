// Package wiring builds runtime dependencies for the CLI runner.
package wiring

import (
	"context"
	"io"
	"net/http"
	"time"

	"stool-grabber/internal/ai"
	"stool-grabber/internal/aiinfra/openrouter"
	"stool-grabber/internal/cli"
	"stool-grabber/internal/telegram"
)

type Deps struct {
	RunnerDeps cli.Deps
}

type Runtime struct {
	Telegram   telegram.Credentials
	OpenRouter openrouter.Config

	In  io.Reader
	Out io.Writer

	HTTPClient        *http.Client
	OpenRouterTimeout time.Duration
}

func NewDeps(rt Runtime) (Deps, error) {
	invoker, err := openrouter.NewInvoker(rt.OpenRouter, openrouter.Options{
		HTTPClient: rt.HTTPClient,
		Timeout:    rt.OpenRouterTimeout,
	})
	if err != nil {
		return Deps{}, err
	}
	prompts, err := ai.NewEmbeddedPrompts()
	if err != nil {
		return Deps{}, err
	}
	analyzer := &ai.Analyzer{Prompts: prompts, Invoker: invoker}

	return Deps{
		RunnerDeps: cli.Deps{
			TelegramCredentials: rt.Telegram,
			Analyzer:            analyzer,
			In:                  rt.In,
			Out:                 rt.Out,
		},
	}, nil
}

func RunJob(ctx context.Context, deps Deps, job *cli.JobConfig) error {
	if job == nil {
		return io.ErrUnexpectedEOF
	}
	return cli.RunJob(ctx, deps.RunnerDeps, job)
}

