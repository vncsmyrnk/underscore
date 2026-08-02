package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func resultError(plan Plan, stage string, kind pipeline.ErrorKind, err error) error {
	return &pipeline.Error{
		Kind:     kind,
		Pipeline: plan.Name,
		Stage:    stage,
		Err:      err,
	}
}

func outcomeError(plan Plan, outcome StageOutcome) error {
	switch {
	case outcome.Err != nil:
		kind := pipeline.ErrLaunch
		if errors.Is(outcome.Err, context.Canceled) {
			kind = pipeline.ErrCancellation
		}
		return resultError(plan, outcome.Name, kind, outcome.Err)
	case outcome.Signal != "":
		return resultError(plan, outcome.Name, pipeline.ErrSignal, errors.New(outcome.Signal))
	case outcome.ExitCode != 0:
		return resultError(plan, outcome.Name, pipeline.ErrExit, fmt.Errorf("exit status %d", outcome.ExitCode))
	default:
		return nil
	}
}
