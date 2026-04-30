package ai

import (
	"context"
	"fmt"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"

	"github.com/skosovsky/prompty"
)

// BuildExecution builds a prompt execution for AnalyzeCore (Task 07).
func BuildExecution(
	ctx context.Context,
	prompts *contractgen.Prompts,
	params AnalyzeParams,
	agg *aggregate.Result,
) (*prompty.PromptExecution, error) {
	if prompts == nil {
		return nil, fmt.Errorf("prompts is nil")
	}
	id, input, err := BuildAnalyzeCoreInput(params, agg)
	if err != nil {
		return nil, err
	}
	if id != contractgen.AnalyzeCore {
		return nil, fmt.Errorf("unexpected prompt id: %s", id)
	}

	input = ApplyBudget(input, DefaultBudgetOptions())

	exec, err := prompts.RenderAnalyzeCore(ctx, input)
	if err != nil {
		return nil, err
	}
	if exec.ModelOptions == nil {
		exec.ModelOptions = &prompty.ModelOptions{}
	}
	exec.ModelOptions.Model = params.Model
	exec.ModelOptions.Temperature = &params.Temperature

	return exec, nil
}

