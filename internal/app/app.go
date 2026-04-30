// Package app wires CLI use-cases (composition root inside internal/).
package app

import (
	"context"

	"stool-grabber/internal/cli"
	"stool-grabber/internal/config"
	"stool-grabber/internal/wiring"
)

func Run(ctx context.Context, configPath string) error {
	job, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	deps, err := wiring.NewDepsFromEnv()
	if err != nil {
		return err
	}
	return wiring.RunJob(ctx, deps, cli.FromJob(job))
}
