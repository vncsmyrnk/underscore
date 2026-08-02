package run

import (
	"context"
	"errors"
	"testing"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func TestOutcomeErrorMapsExitFailureToPipelineError(t *testing.T) {
	t.Parallel()

	plan := Plan{Name: "workflow-v1"}
	outcome := StageOutcome{
		Name:     "requisite",
		Kind:     StageKindRequisite,
		ExitCode: 2,
	}

	err := outcomeError(plan, outcome)
	if err == nil {
		t.Fatal("expected exit failure")
	}

	var pipelineErr *pipeline.Error
	if !errors.As(err, &pipelineErr) {
		t.Fatalf("expected pipeline.Error, got %T", err)
	}

	if pipelineErr.Kind != pipeline.ErrExit {
		t.Fatalf("kind = %q, want %q", pipelineErr.Kind, pipeline.ErrExit)
	}

	if pipelineErr.Stage != "requisite" {
		t.Fatalf("stage = %q, want requisite", pipelineErr.Stage)
	}
}

func TestOutcomeErrorMapsSignalFailureToPipelineError(t *testing.T) {
	t.Parallel()

	err := outcomeError(Plan{Name: "workflow-v1"}, StageOutcome{
		Name:   "pipeline",
		Kind:   StageKindPipeline,
		Signal: "SIGTERM",
	})
	if err == nil {
		t.Fatal("expected signal failure")
	}

	var pipelineErr *pipeline.Error
	if !errors.As(err, &pipelineErr) {
		t.Fatalf("expected pipeline.Error, got %T", err)
	}

	if pipelineErr.Kind != pipeline.ErrSignal {
		t.Fatalf("kind = %q, want %q", pipelineErr.Kind, pipeline.ErrSignal)
	}
}

func TestOutcomeErrorMapsCancellationFailureToPipelineError(t *testing.T) {
	t.Parallel()

	err := outcomeError(Plan{Name: "workflow-v1"}, StageOutcome{
		Name: "pipeline",
		Kind: StageKindPipeline,
		Err:  context.Canceled,
	})
	if err == nil {
		t.Fatal("expected cancellation failure")
	}

	var pipelineErr *pipeline.Error
	if !errors.As(err, &pipelineErr) {
		t.Fatalf("expected pipeline.Error, got %T", err)
	}

	if pipelineErr.Kind != pipeline.ErrCancellation {
		t.Fatalf("kind = %q, want %q", pipelineErr.Kind, pipeline.ErrCancellation)
	}
}
