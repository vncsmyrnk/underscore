package run

import (
	"context"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

type Service struct {
	runner ProcessRunner
	writer EffectWriter
}

func NewService(runner ProcessRunner, writer EffectWriter) Service {
	return Service{
		runner: runner,
		writer: writer,
	}
}

func (s Service) Run(ctx context.Context, plan Plan) (ExecutionResult, error) {
	if plan.Pipeline.Requisite != nil {
		result, err := s.runner.Run(ctx, requisitePlan(plan))
		if err != nil {
			return ExecutionResult{}, resultError(plan, "requisite", pipeline.ErrLaunch, err)
		}

		if err := evaluateRequisite(plan, result); err != nil {
			return ExecutionResult{}, err
		}
	}

	actionResult, err := s.runner.Run(ctx, actionPlan(plan))
	if err != nil {
		return ExecutionResult{}, resultError(plan, "pipeline", pipeline.ErrLaunch, err)
	}

	actionResult, err = evaluateAction(plan, actionResult)
	if err != nil {
		return ExecutionResult{}, err
	}

	if plan.Pipeline.Afterwards != nil {
		result, err := s.runner.Run(ctx, afterwardsPlan(plan))
		if err != nil {
			return ExecutionResult{}, resultError(plan, "afterwards", pipeline.ErrLaunch, err)
		}

		if err := evaluateAfterwards(plan, result); err != nil {
			return ExecutionResult{}, err
		}
	}

	if actionResult.Effect != nil {
		if err := s.writer.Write(ctx, *actionResult.Effect); err != nil {
			return ExecutionResult{}, resultError(plan, "effect", pipeline.ErrEffectIPC, err)
		}
	}

	return actionResult, nil
}
