package run

import (
	"errors"
	"testing"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func TestActionPlanDropsLifecycleSteps(t *testing.T) {
	t.Parallel()

	plan := Plan{
		Pipeline: pipeline.Pipeline{
			Requisite:  &pipeline.Step{Role: pipeline.RoleRequisite},
			Source:     &pipeline.Step{Role: pipeline.RoleSource},
			Transforms: []pipeline.Step{{Role: pipeline.RoleTransform}},
			Command:    &pipeline.Command{Argv: []string{"cat"}},
			Effect:     pipeline.EffectCD,
			Afterwards: &pipeline.Step{Role: pipeline.RoleAfterwards},
		},
	}

	got := actionPlan(plan)

	if got.Pipeline.Requisite != nil || got.Pipeline.Afterwards != nil {
		t.Fatalf("unexpected action-only pipeline: %#v", got.Pipeline)
	}

	if got.Pipeline.Source == nil || len(got.Pipeline.Transforms) != 1 || got.Pipeline.Command == nil {
		t.Fatalf("expected streamed action to be preserved: %#v", got.Pipeline)
	}
}

func TestFirstPipelineFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		outcomes []StageOutcome
		wantName string
		wantNil  bool
	}{
		{
			name: "earliest ordinary failure wins",
			outcomes: []StageOutcome{
				{Name: "source", Kind: StageKindPipeline, ExitCode: 3},
				{Name: "transform", Kind: StageKindPipeline, ExitCode: 7},
			},
			wantName: "source",
		},
		{
			name: "expected sigpipe is ignored when downstream succeeds",
			outcomes: []StageOutcome{
				{Name: "source", Kind: StageKindPipeline, Signal: "SIGPIPE"},
				{Name: "transform", Kind: StageKindPipeline, ExitCode: 0},
			},
			wantNil: true,
		},
		{
			name: "sigpipe is not ignored when downstream fails",
			outcomes: []StageOutcome{
				{Name: "source", Kind: StageKindPipeline, Signal: "SIGPIPE"},
				{Name: "transform", Kind: StageKindPipeline, ExitCode: 9},
			},
			wantName: "source",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := firstPipelineFailure(tt.outcomes)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil failure, got %#v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected failure")
			}

			if got.Name != tt.wantName {
				t.Fatalf("name = %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

func TestEvaluateActionReturnsStdoutForSuccessfulNonEffectPipeline(t *testing.T) {
	t.Parallel()

	result, err := evaluateAction(Plan{
		Name: "workflow-v1",
		Pipeline: pipeline.Pipeline{
			Source: &pipeline.Step{Role: pipeline.RoleSource},
		},
	}, ExecutionResult{
		Outcomes: []StageOutcome{{Name: "pipeline", Kind: StageKindPipeline}},
		Stdout:   []byte("hello\n"),
	})
	if err != nil {
		t.Fatalf("evaluateAction() error = %v", err)
	}

	if string(result.Stdout) != "hello\n" {
		t.Fatalf("stdout = %q, want hello\\n", result.Stdout)
	}

	if result.Effect != nil {
		t.Fatalf("expected no effect, got %#v", result.Effect)
	}
}

func TestEvaluateActionRejectsInvalidEffectTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		stdout []byte
	}{
		{name: "empty first line", stdout: []byte("\nignored\n")},
		{name: "nul byte", stdout: []byte("/tmp/demo\x00\nignored\n")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := evaluateAction(Plan{
				Name: "workflow-v1",
				Pipeline: pipeline.Pipeline{
					Effect: pipeline.EffectCD,
				},
			}, ExecutionResult{
				Outcomes: []StageOutcome{{Name: "pipeline", Kind: StageKindPipeline}},
				Stdout:   tt.stdout,
			})
			if err == nil {
				t.Fatal("expected effect validation failure")
			}

			var pipelineErr *pipeline.Error
			if !errors.As(err, &pipelineErr) {
				t.Fatalf("expected pipeline.Error, got %T", err)
			}

			if pipelineErr.Kind != pipeline.ErrEffectValidation {
				t.Fatalf("kind = %q, want %q", pipelineErr.Kind, pipeline.ErrEffectValidation)
			}
		})
	}
}

func TestEvaluateActionExtractsFirstLineEffect(t *testing.T) {
	t.Parallel()

	result, err := evaluateAction(Plan{
		Name: "workflow-v1",
		Pipeline: pipeline.Pipeline{
			Effect: pipeline.EffectCD,
		},
	}, ExecutionResult{
		Outcomes: []StageOutcome{{Name: "pipeline", Kind: StageKindPipeline}},
		Stdout:   []byte("/tmp/demo\nignored\n"),
	})
	if err != nil {
		t.Fatalf("evaluateAction() error = %v", err)
	}

	if result.Effect == nil {
		t.Fatal("expected effect result")
	}

	if result.Effect.Value != "/tmp/demo" {
		t.Fatalf("effect value = %q, want /tmp/demo", result.Effect.Value)
	}

	if len(result.Stdout) != 0 {
		t.Fatalf("stdout = %q, want empty", result.Stdout)
	}
}
