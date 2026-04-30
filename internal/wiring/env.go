package wiring

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stool-grabber/internal/aiinfra/openrouter"
	"stool-grabber/internal/telegram"
)

func NewDepsFromEnv() (Deps, error) {
	tgCreds, err := telegram.NewCredentials(
		os.Getenv("TG_APP_ID"),
		os.Getenv("TG_APP_HASH"),
		absIfRelative(strings.TrimSpace(os.Getenv("TG_SESSION_PATH"))),
	)
	if err != nil {
		return Deps{}, err
	}

	orCfg, err := openrouter.NewConfig(
		strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")),
		strings.TrimSpace(os.Getenv("OPENROUTER_BASE_URL")),
	)
	if err != nil {
		return Deps{}, err
	}

	// Explicit HTTP client from composition root (runtime config).
	httpClient := &http.Client{}

	return NewDeps(Runtime{
		Telegram:   tgCreds,
		OpenRouter: orCfg,
		In:         os.Stdin,
		Out:        os.Stdout,
		HTTPClient: httpClient,
		OpenRouterTimeout: 300 * time.Second,
	})
}

func absIfRelative(p string) string {
	if strings.TrimSpace(p) == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return p
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		// Best-effort; caller will still use the relative path.
		return p
	}
	return abs
}

