package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"stool-grabber/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: stool-grabber run -c <job.yaml>")
	}

	switch args[0] {
	case "run":
		fs := flag.NewFlagSet("run", flag.ContinueOnError)
		configPath := fs.String("c", "", "path to job.yaml")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("load .env: %w", err)
		}

		return app.Run(ctx, *configPath)
	default:
		return fmt.Errorf("unknown command %q, expected: run", args[0])
	}
}
