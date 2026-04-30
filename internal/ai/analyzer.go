package ai

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"

	"github.com/skosovsky/prompty"
	"github.com/skosovsky/routery"
)

type Analyzer struct {
	Prompts *contractgen.Prompts
	Invoker prompty.Invoker
}

func (a *Analyzer) Analyze(ctx context.Context, params AnalyzeParams, agg *aggregate.Result) (*contractgen.AnalyzeCoreOutput, error) {
	if a == nil || a.Prompts == nil || a.Invoker == nil {
		return nil, fmt.Errorf("analyze: analyzer is not configured")
	}
	if agg == nil {
		return nil, fmt.Errorf("analyze: agg is nil")
	}
	exec, err := BuildExecution(ctx, a.Prompts, params, agg)
	if err != nil {
		return nil, err
	}
	return analyzeWithInvoker(ctx, a.Invoker, exec)
}

func analyzeWithInvoker(ctx context.Context, inv prompty.Invoker, exec *prompty.PromptExecution) (*contractgen.AnalyzeCoreOutput, error) {
	base := routery.ExecutorFunc[*prompty.PromptExecution, *contractgen.AnalyzeCoreOutput](func(c context.Context, e *prompty.PromptExecution) (*contractgen.AnalyzeCoreOutput, error) {
		out, err := prompty.ExecuteWithStructuredOutput[contractgen.AnalyzeCoreOutput](c, inv, e)
		if err != nil {
			return nil, err
		}
		return out, nil
	})

	pred := func(_ context.Context, _ *prompty.PromptExecution, err error) bool {
		if err == nil {
			return false
		}
		// Never retry deterministic structured-output validation errors.
		var ve *prompty.ValidationError
		if errors.As(err, &ve) {
			return false
		}
		if ctx.Err() != nil {
			return false
		}
		var ne net.Error
		if errors.As(err, &ne) {
			return ne.Timeout() || ne.Temporary()
		}
		return err == context.DeadlineExceeded
	}

	execWithMW := routery.Apply(
		base,
		routery.Timeout[*prompty.PromptExecution, *contractgen.AnalyzeCoreOutput](90*time.Second),
		routery.RetryIf[*prompty.PromptExecution, *contractgen.AnalyzeCoreOutput](3, 1500*time.Millisecond, pred),
	)

	return execWithMW.Execute(ctx, exec)
}

