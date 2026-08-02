package run

import (
	"bytes"
	"fmt"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func actionPlan(plan Plan) Plan {
	stage := plan
	stage.Pipeline.Requisite = nil
	stage.Pipeline.Afterwards = nil
	return stage
}

func evaluateAction(plan Plan, result ExecutionResult) (ExecutionResult, error) {
	if failure := firstPipelineFailure(result.Outcomes); failure != nil {
		return ExecutionResult{}, outcomeError(plan, *failure)
	}

	if plan.Pipeline.Effect == "" {
		return result, nil
	}

	value, err := extractEffectValue(result.Stdout)
	if err != nil {
		return ExecutionResult{}, resultError(plan, "pipeline", pipeline.ErrEffectValidation, err)
	}

	result.Stdout = nil
	result.Effect = &EffectResult{
		Name:  plan.Pipeline.Effect,
		Value: value,
	}

	return result, nil
}

func firstPipelineFailure(outcomes []StageOutcome) *StageOutcome {
	for i := range outcomes {
		outcome := outcomes[i]
		switch {
		case outcome.Err != nil:
			return &outcomes[i]
		case outcome.Signal != "":
			if outcome.Signal == "SIGPIPE" && downstreamSucceeded(outcomes[i+1:]) {
				continue
			}
			return &outcomes[i]
		case outcome.ExitCode != 0:
			return &outcomes[i]
		}
	}

	return nil
}

func downstreamSucceeded(outcomes []StageOutcome) bool {
	if len(outcomes) == 0 {
		return false
	}

	for _, outcome := range outcomes {
		if outcome.Err != nil || outcome.Signal != "" || outcome.ExitCode != 0 {
			return false
		}
	}

	return true
}

func extractEffectValue(stdout []byte) (string, error) {
	line := stdout
	if index := bytes.IndexByte(stdout, '\n'); index >= 0 {
		line = stdout[:index]
	}

	line = bytes.TrimSuffix(line, []byte{'\r'})
	if len(line) == 0 {
		return "", fmt.Errorf("effect target is empty")
	}

	if bytes.IndexByte(line, 0) >= 0 {
		return "", fmt.Errorf("effect target contains NUL")
	}

	return string(line), nil
}
