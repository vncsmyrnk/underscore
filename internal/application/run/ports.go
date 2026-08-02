package run

import (
	"context"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

type ConfigLoader interface {
	LoadPipeline(ctx context.Context, name string) (pipeline.Pipeline, error)
	LoadProfile(ctx context.Context, pipelineName string, profileName string) (map[string]string, error)
}

type ProcessRunner interface {
	Run(ctx context.Context, plan Plan) (ExecutionResult, error)
}

type EffectWriter interface {
	Write(ctx context.Context, result EffectResult) error
}

type Plan struct {
	Name      string
	Profile   string
	Pipeline  pipeline.Pipeline
	Overrides map[string]string
	Resolved  map[string]pipeline.ResolvedValue
}
