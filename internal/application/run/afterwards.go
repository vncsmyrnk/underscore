package run

import (
	"fmt"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func afterwardsPlan(plan Plan) Plan {
	stage := plan
	stage.Pipeline.Requisite = nil
	stage.Pipeline.Source = nil
	stage.Pipeline.Transforms = nil
	stage.Pipeline.Command = nil
	stage.Pipeline.Effect = ""
	return stage
}

func evaluateAfterwards(plan Plan, result ExecutionResult) error {
	if len(result.Outcomes) == 0 {
		return resultError(plan, "afterwards", pipeline.ErrExit, fmt.Errorf("missing afterwards outcome"))
	}

	return outcomeError(plan, result.Outcomes[0])
}
