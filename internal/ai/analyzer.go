package ai

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"

	"github.com/skosovsky/prompty"
	"github.com/skosovsky/routery"
)

type Analyzer struct {
	Prompts *contractgen.Prompts
	Invoker prompty.Invoker
	Timeout time.Duration
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
	timeout := a.Timeout
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	return analyzeWithInvoker(ctx, a.Invoker, exec, timeout)
}

func analyzeWithInvoker(ctx context.Context, inv prompty.Invoker, exec *prompty.PromptExecution, timeout time.Duration) (*contractgen.AnalyzeCoreOutput, error) {
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
		// Never retry deterministic structured-output validation errors,
		// except for common truncation cases (e.g. incomplete JSON).
		var ve *prompty.ValidationError
		if errors.As(err, &ve) {
			if ve != nil && ve.Err != nil {
				msg := ve.Err.Error()
				// Truncated JSON is often transient on overloaded/free endpoints.
				if strings.Contains(msg, "unexpected end of JSON input") ||
					strings.Contains(msg, "unexpected EOF") {
					return true
				}
			}
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
		routery.RetryIf[*prompty.PromptExecution, *contractgen.AnalyzeCoreOutput](5, 1500*time.Millisecond, pred),
		routery.Timeout[*prompty.PromptExecution, *contractgen.AnalyzeCoreOutput](timeout),
	)

	return execWithMW.Execute(ctx, exec)
}

