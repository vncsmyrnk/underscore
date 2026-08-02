package run

import (
	"fmt"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func requisitePlan(plan Plan) Plan {
	stage := plan
	stage.Pipeline.Source = nil
	stage.Pipeline.Transforms = nil
	stage.Pipeline.Command = nil
	stage.Pipeline.Effect = ""
	stage.Pipeline.Afterwards = nil
	return stage
}

func evaluateRequisite(plan Plan, result ExecutionResult) error {
	if len(result.Outcomes) == 0 {
		return resultError(plan, "requisite", pipeline.ErrExit, fmt.Errorf("missing requisite outcome"))
	}

	outcome := result.Outcomes[0]
	if outcome.Err != nil || outcome.Signal != "" {
		return outcomeError(plan, outcome)
	}

	invert := plan.Pipeline.Requisite != nil && plan.Pipeline.Requisite.Invert
	if invert {
		if outcome.ExitCode != 0 {
			return nil
		}
		return resultError(plan, outcome.Name, pipeline.ErrExit, fmt.Errorf("exit status %d", outcome.ExitCode))
	}

	return outcomeError(plan, outcome)
}
