package run

import (
	"context"
	"errors"
	"testing"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

type stubRunner struct {
	results []ExecutionResult
	errs    []error
	calls   []Plan
}

func (s *stubRunner) Run(_ context.Context, plan Plan) (ExecutionResult, error) {
	s.calls = append(s.calls, plan)

	var result ExecutionResult
	if len(s.results) > 0 {
		result = s.results[0]
		s.results = s.results[1:]
	}

	var err error
	if len(s.errs) > 0 {
		err = s.errs[0]
		s.errs = s.errs[1:]
	}

	return result, err
}

type stubWriter struct {
	writes []EffectResult
	err    error
}

func (s *stubWriter) Write(_ context.Context, result EffectResult) error {
	s.writes = append(s.writes, result)
	return s.err
}

func TestServiceRunStopsAfterFailingRequisite(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		results: []ExecutionResult{
			{Outcomes: []StageOutcome{{Name: "requisite", Kind: StageKindRequisite, ExitCode: 2}}},
		},
	}

	service := NewService(runner, &stubWriter{})

	_, err := service.Run(context.Background(), Plan{
		Name: "workflow-v1",
		Pipeline: pipeline.Pipeline{
			Requisite: &pipeline.Step{Role: pipeline.RoleRequisite},
			Source:    &pipeline.Step{Role: pipeline.RoleSource},
		},
	})
	if err == nil {
		t.Fatal("expected requisite failure")
	}

	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(runner.calls))
	}

	if runner.calls[0].Pipeline.Source != nil {
		t.Fatalf("expected requisite subplan, got %#v", runner.calls[0].Pipeline)
	}
}

func TestServiceRunSuppressesEffectAfterAfterwardsFailure(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		results: []ExecutionResult{
			{Outcomes: []StageOutcome{{Name: "pipeline", Kind: StageKindPipeline}}, Stdout: []byte("/tmp/demo\n")},
			{Outcomes: []StageOutcome{{Name: "afterwards", Kind: StageKindAfterwards, ExitCode: 3}}},
		},
	}
	writer := &stubWriter{}

	service := NewService(runner, writer)

	_, err := service.Run(context.Background(), Plan{
		Name: "workflow-v1",
		Pipeline: pipeline.Pipeline{
			Source:     &pipeline.Step{Role: pipeline.RoleSource},
			Effect:     pipeline.EffectCD,
			Afterwards: &pipeline.Step{Role: pipeline.RoleAfterwards},
		},
	})
	if err == nil {
		t.Fatal("expected afterwards failure")
	}

	if len(writer.writes) != 0 {
		t.Fatalf("writer writes = %#v, want none", writer.writes)
	}
}

func TestServiceRunWritesValidatedEffectOnSuccess(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		results: []ExecutionResult{
			{Outcomes: []StageOutcome{{Name: "pipeline", Kind: StageKindPipeline}}, Stdout: []byte("/tmp/demo\nignored\n")},
		},
	}
	writer := &stubWriter{}

	service := NewService(runner, writer)

	result, err := service.Run(context.Background(), Plan{
		Name: "workflow-v1",
		Pipeline: pipeline.Pipeline{
			Source: &pipeline.Step{Role: pipeline.RoleSource},
			Effect: pipeline.EffectCD,
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Effect == nil || result.Effect.Value != "/tmp/demo" {
		t.Fatalf("unexpected effect result: %#v", result.Effect)
	}

	if len(writer.writes) != 1 || writer.writes[0].Value != "/tmp/demo" {
		t.Fatalf("writer writes = %#v", writer.writes)
	}
}

func TestServiceRunReturnsEffectIPCFailure(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		results: []ExecutionResult{
			{Outcomes: []StageOutcome{{Name: "pipeline", Kind: StageKindPipeline}}, Stdout: []byte("/tmp/demo\n")},
		},
	}
	writer := &stubWriter{err: errors.New("effect write failed")}

	service := NewService(runner, writer)

	_, err := service.Run(context.Background(), Plan{
		Name: "workflow-v1",
		Pipeline: pipeline.Pipeline{
			Source: &pipeline.Step{Role: pipeline.RoleSource},
			Effect: pipeline.EffectCD,
		},
	})
	if err == nil {
		t.Fatal("expected effect write failure")
	}

	var pipelineErr *pipeline.Error
	if !errors.As(err, &pipelineErr) {
		t.Fatalf("expected pipeline.Error, got %T", err)
	}

	if pipelineErr.Kind != pipeline.ErrEffectIPC {
		t.Fatalf("kind = %q, want %q", pipelineErr.Kind, pipeline.ErrEffectIPC)
	}
}

func TestServiceRunReturnsStdoutForSuccessfulNonEffectPipeline(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		results: []ExecutionResult{
			{Outcomes: []StageOutcome{{Name: "pipeline", Kind: StageKindPipeline}}, Stdout: []byte("hello\n")},
		},
	}

	service := NewService(runner, &stubWriter{})

	result, err := service.Run(context.Background(), Plan{
		Name: "workflow-v1",
		Pipeline: pipeline.Pipeline{
			Source: &pipeline.Step{Role: pipeline.RoleSource},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if string(result.Stdout) != "hello\n" {
		t.Fatalf("stdout = %q, want hello\\n", result.Stdout)
	}
}
