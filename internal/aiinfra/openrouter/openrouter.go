// Package openrouter provides a prompty.Invoker backed by OpenRouter (OpenAI-compatible API).
package openrouter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	openaiadapter "github.com/skosovsky/prompty/adapter/openai"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/skosovsky/prompty"
	"github.com/skosovsky/prompty/adapter"
)

const DefaultBaseURL = "https://openrouter.ai/api/v1"

type Config struct {
	APIKey  string
	BaseURL string
}

func NewConfig(apiKey string, baseURL string) (Config, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return Config{}, fmt.Errorf("OPENROUTER_API_KEY is required")
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return Config{APIKey: apiKey, BaseURL: baseURL}, nil
}

type Options struct {
	HTTPClient *http.Client
	Timeout    time.Duration
}

func NewInvoker(cfg Config, opt Options) (prompty.Invoker, error) {
	httpClient := opt.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	if httpClient.Timeout == 0 {
		if opt.Timeout > 0 {
			httpClient.Timeout = opt.Timeout
		} else {
			httpClient.Timeout = 120 * time.Second
		}
	}

	sdk := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
		option.WithHTTPClient(httpClient),
	)

	adp := openaiadapter.New(openaiadapter.WithClient(&sdk))
	return adapter.NewClient[*openai.ChatCompletionNewParams, *openai.ChatCompletion](adp), nil
}

