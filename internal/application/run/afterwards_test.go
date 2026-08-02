package run

import (
	"errors"
	"testing"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func TestAfterwardsPlanKeepsOnlyAfterwardsStep(t *testing.T) {
	t.Parallel()

	plan := Plan{
		Pipeline: pipeline.Pipeline{
			Source:     &pipeline.Step{Role: pipeline.RoleSource},
			Transforms: []pipeline.Step{{Role: pipeline.RoleTransform}},
			Command:    &pipeline.Command{Argv: []string{"cat"}},
			Effect:     pipeline.EffectCD,
			Afterwards: &pipeline.Step{Role: pipeline.RoleAfterwards},
		},
	}

	got := afterwardsPlan(plan)

	if got.Pipeline.Afterwards == nil {
		t.Fatal("expected afterwards to be preserved")
	}

	if got.Pipeline.Source != nil || len(got.Pipeline.Transforms) != 0 || got.Pipeline.Command != nil || got.Pipeline.Effect != "" {
		t.Fatalf("unexpected afterwards-only pipeline: %#v", got.Pipeline)
	}
}

func TestEvaluateAfterwards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result ExecutionResult
		want   pipeline.ErrorKind
	}{
		{
			name: "ordinary success passes",
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "afterwards", Kind: StageKindAfterwards}},
			},
		},
		{
			name: "ordinary failure fails",
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "afterwards", Kind: StageKindAfterwards, ExitCode: 9}},
			},
			want: pipeline.ErrExit,
		},
		{
			name: "signal failure fails",
			result: ExecutionResult{
				Outcomes: []StageOutcome{{Name: "afterwards", Kind: StageKindAfterwards, Signal: "SIGTERM"}},
			},
			want: pipeline.ErrSignal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := evaluateAfterwards(Plan{Name: "workflow-v1"}, tt.result)
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
