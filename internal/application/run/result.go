package run

import "github.com/vncsmyrnk/underscore/internal/core/pipeline"

type StageKind string

const (
	StageKindRequisite  StageKind = "requisite"
	StageKindPipeline   StageKind = "pipeline"
	StageKindAfterwards StageKind = "afterwards"
)

type StageOutcome struct {
	Name     string
	Kind     StageKind
	ExitCode int
	Signal   string
	Err      error
}

type EffectResult struct {
	Name  pipeline.EffectName
	Value string
}

type ExecutionResult struct {
	Outcomes []StageOutcome
	Stdout   []byte
	Effect   *EffectResult
}
