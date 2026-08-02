package run

import (
	"context"
	"errors"
	"testing"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func TestRequisitePlanKeepsOnlyRequisiteStep(t *testing.T) {
	t.Parallel()

	plan := Plan{
		Pipeline: pipeline.Pipeline{
			Requisite: &pipeline.Step{Role: pipeline.RoleRequisite},
			Source:    &pipeline.Step{Role: pipeline.RoleSource},
			Transforms: []pipeline.Step{
				{Role: pipeline.RoleTransform},
			},
			Command:    &pipeline.Command{Argv: []string{"cat"}},
			Effect:     pipeline.EffectCD,
			Afterwards: &pipeline.Step{Role: pipeline.RoleAfterwards},
		},
	}

	got := requisitePlan(plan)

	if got.Pipeline.Requisite == nil {
		t.Fatal("expected requisite to be preserved")
	}

	if got.Pipeline.Source != nil || len(got.Pipeline.Transforms) != 0 || got.Pipeline.Command != nil || got.Pipeline.Effect != "" || got.Pipeline.Afterwards != nil {
		t.Fatalf("unexpected requisite-only pipeline: %#v", got.Pipeline)
	}
}

func TestEvaluateRequisite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		step   pipeline.Step
		result ExecutionResult
		want   pipeline.ErrorKind
	}{
		{
			name: "ordinary success without inversion passes",
			step: pipeline.Step{Role: pipeline.RoleRequisite},
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "requisite", Kind: StageKindRequisite, ExitCode: 0}},
			},
		},
		{
			name: "ordinary failure without inversion fails",
			step: pipeline.Step{Role: pipeline.RoleRequisite},
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "requisite", Kind: StageKindRequisite, ExitCode: 7}},
			},
			want: pipeline.ErrExit,
		},
		{
			name: "ordinary success with inversion fails",
			step: pipeline.Step{Role: pipeline.RoleRequisite, Invert: true},
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "requisite", Kind: StageKindRequisite, ExitCode: 0}},
			},
			want: pipeline.ErrExit,
		},
		{
			name: "ordinary failure with inversion passes",
			step: pipeline.Step{Role: pipeline.RoleRequisite, Invert: true},
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "requisite", Kind: StageKindRequisite, ExitCode: 7}},
			},
		},
		{
			name: "cancellation is never inverted",
			step: pipeline.Step{Role: pipeline.RoleRequisite, Invert: true},
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "requisite", Kind: StageKindRequisite, Err: context.Canceled}},
			},
			want: pipeline.ErrCancellation,
		},
		{
			name: "signals are never inverted",
			step: pipeline.Step{Role: pipeline.RoleRequisite, Invert: true},
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "requisite", Kind: StageKindRequisite, Signal: "SIGTERM"}},
			},
			want: pipeline.ErrSignal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := evaluateRequisite(Plan{
				Name: "workflow-v1",
				Pipeline: pipeline.Pipeline{
					Requisite: &tt.step,
				},
			}, tt.result)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected %q failure", tt.want)
			}

			var pipelineErr *pipeline.Error
			if !errors.As(err, &pipelineErr) {
				t.Fatalf("expected pipeline.Error, got %T", err)
			}

			if pipelineErr.Kind != tt.want {
				t.Fatalf("kind = %q, want %q", pipelineErr.Kind, tt.want)
			}
		})
	}
}
