package run

import (
	"context"
	"testing"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

type fakeConfigLoader struct{}

func (fakeConfigLoader) LoadPipeline(context.Context, string) (pipeline.Pipeline, error) {
	return pipeline.Pipeline{}, nil
}

func (fakeConfigLoader) LoadProfile(context.Context, string, string) (map[string]string, error) {
	return map[string]string{}, nil
}

type fakeProcessRunner struct{}

func (fakeProcessRunner) Run(context.Context, Plan) (ExecutionResult, error) {
	return ExecutionResult{}, nil
}

type fakeEffectWriter struct{}

func (fakeEffectWriter) Write(context.Context, EffectResult) error {
	return nil
}

var (
	_ ConfigLoader  = fakeConfigLoader{}
	_ ProcessRunner = fakeProcessRunner{}
	_ EffectWriter  = fakeEffectWriter{}
)

func TestPlanCarriesPipelineAndResolutionData(t *testing.T) {
	t.Parallel()

	plan := Plan{
		Name:    "workflow-v1",
		Profile: "default",
		Pipeline: pipeline.Pipeline{
			Description: "choose a worktree",
			Effect:      pipeline.EffectCD,
		},
		Overrides: map[string]string{
			"target-path": "/tmp/demo",
		},
		Resolved: map[string]pipeline.ResolvedValue{
			"TARGET_PATH": {
				Name:  "TARGET_PATH",
				Value: "/tmp/demo",
			},
		},
	}

	if plan.Pipeline.Effect != pipeline.EffectCD {
		t.Fatalf("effect = %q, want %q", plan.Pipeline.Effect, pipeline.EffectCD)
	}

	if plan.Resolved["TARGET_PATH"].Value != "/tmp/demo" {
		t.Fatalf("resolved override = %q", plan.Resolved["TARGET_PATH"].Value)
	}
}

func TestExecutionResultCarriesOneOutcomePerStage(t *testing.T) {
	t.Parallel()

	result := ExecutionResult{
		Outcomes: []StageOutcome{
			{Name: "requisite", Kind: StageKindRequisite, ExitCode: 0},
			{Name: "pipeline", Kind: StageKindPipeline, ExitCode: 0},
			{Name: "afterwards", Kind: StageKindAfterwards, ExitCode: 0},
		},
		Stdout: []byte("output"),
		Effect: &EffectResult{
			Name:  pipeline.EffectCD,
			Value: "/tmp/worktree",
		},
	}

	if len(result.Outcomes) != 3 {
		t.Fatalf("expected 3 stage outcomes, got %d", len(result.Outcomes))
	}

	if result.Effect == nil || result.Effect.Value != "/tmp/worktree" {
		t.Fatalf("unexpected effect result: %#v", result.Effect)
	}
}

func TestPortsAreSatisfiedByConsumerOwnedFakes(t *testing.T) {
	t.Parallel()
}
