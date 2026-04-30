package ai

import (
	"embed"
	"fmt"

	"stool-grabber/internal/ai/contractgen"

	"github.com/skosovsky/prompty/embedregistry"
	"github.com/skosovsky/prompty/parser/yaml"
)

//go:embed prompts/**
var promptsFS embed.FS

func NewEmbeddedPrompts() (*contractgen.Prompts, error) {
	reg, err := embedregistry.New(promptsFS, "prompts", embedregistry.WithParser(yaml.New()))
	if err != nil {
		return nil, fmt.Errorf("create embedded prompt registry: %w", err)
	}
	return contractgen.NewPrompts(reg), nil
}

