package pipeline

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStepRoleConstants(t *testing.T) {
	t.Parallel()

	got := []StepRole{
		RoleRequisite,
		RoleSource,
		RoleTransform,
		RoleAfterwards,
	}

	want := []StepRole{
		"requisite",
		"source",
		"transform",
		"afterwards",
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("role %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCommandPreservesArgvAndBecome(t *testing.T) {
	t.Parallel()

	command := Command{
		Argv:   []string{"mount", "/dev/mapper/demo", "/mnt/demo"},
		Become: true,
	}

	if !command.Become {
		t.Fatal("expected become flag to be preserved")
	}

	if len(command.Argv) != 3 {
		t.Fatalf("expected 3 argv entries, got %d", len(command.Argv))
	}

	if command.Argv[1] != "/dev/mapper/demo" {
		t.Fatalf("argv[1] = %q, want /dev/mapper/demo", command.Argv[1])
	}
}

func TestPipelineRepresentsFrozenNormalizedShape(t *testing.T) {
	t.Parallel()

	pipeline := Pipeline{
		Description: "mount the mapped device",
		Defaults: map[string]string{
			"SRC": "/dev/loop0",
		},
		Trusted: []string{"COMMAND"},
		Resolved: map[string]ResolvedValue{
			"SRC": {
				Name:  "SRC",
				Value: "/dev/loop0",
			},
		},
		Requisite: &Step{
			Role: RoleRequisite,
			Command: Command{
				Argv: []string{"cryptsetup", "open", "$SRC", "vault"},
			},
		},
		Source: &Step{
			Role: RoleSource,
			Command: Command{
				Argv: []string{"git", "worktree", "list"},
			},
		},
		Transforms: []Step{
			{
				Role: RoleTransform,
				Command: Command{
					Argv: []string{"fzf"},
				},
			},
		},
		Command: &Command{
			Argv: []string{"mount", "/dev/mapper/vault", "/mnt/vault"},
		},
		Effect: EffectCD,
		Afterwards: &Step{
			Role: RoleAfterwards,
			Command: Command{
				Argv: []string{"cryptsetup", "close", "vault"},
			},
		},
	}

	if pipeline.Effect != EffectCD {
		t.Fatalf("effect = %q, want %q", pipeline.Effect, EffectCD)
	}

	if pipeline.Requisite == nil || pipeline.Requisite.Role != RoleRequisite {
		t.Fatalf("unexpected requisite: %#v", pipeline.Requisite)
	}

	if len(pipeline.Transforms) != 1 || pipeline.Transforms[0].Role != RoleTransform {
		t.Fatalf("unexpected transforms: %#v", pipeline.Transforms)
	}
}

func TestFrameworkConstants(t *testing.T) {
	t.Parallel()

	if DefaultMaxTransforms != 4 {
		t.Fatalf("DefaultMaxTransforms = %d, want 4", DefaultMaxTransforms)
	}

	if EnvMaxTransforms != "UNDERSCORE_MAX_TRANSFORMS" {
		t.Fatalf("EnvMaxTransforms = %q", EnvMaxTransforms)
	}

	if EnvElevationExecutable != "UNDERSCORE_ELEVATION_EXECUTABLE" {
		t.Fatalf("EnvElevationExecutable = %q", EnvElevationExecutable)
	}

	if EnvEffectResult != "UNDERSCORE_EFFECT_RESULT" {
		t.Fatalf("EnvEffectResult = %q", EnvEffectResult)
	}
}

func TestErrorIncludesContextAndWrapsCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("invalid field")
	err := &Error{
		Kind:     ErrValidation,
		Pipeline: "workflow-v1",
		Field:    "command",
		Stage:    "normalize",
		Err:      cause,
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected wrapped cause to be discoverable with errors.Is")
	}

	message := err.Error()
	for _, want := range []string{"validation", "workflow-v1", "command", "normalize", "invalid field"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected %q in error message %q", want, message)
		}
	}
}

func TestNilErrorReturnsEmptyMessageAndNoCause(t *testing.T) {
	t.Parallel()

	var err *Error
	if err.Error() != "" {
		t.Fatalf("nil error message = %q, want empty string", err.Error())
	}

	if err.Unwrap() != nil {
		t.Fatalf("nil error unwrap = %v, want nil", err.Unwrap())
	}
}

func TestCanonicalFixturesAreStrictJSON(t *testing.T) {
	t.Parallel()

	fixtures := []string{
		"workflow-v1.json",
		"workflow-v1-filter.json",
		"workflow-v1-command-check-and-afterwards.json",
		"workflow-v1-command.json",
		"workflow-v1-no-steps-only.json",
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join("testdata", fixture)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%q) returned error: %v", path, err)
			}

			if !json.Valid(data) {
				t.Fatalf("%q is not valid JSON", path)
			}

			if strings.Contains(string(data), "//") {
				t.Fatalf("%q contains comment syntax", path)
			}

			var decoded map[string]any
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal(%q) returned error: %v", path, err)
			}

			if decoded["version"] != float64(1) {
				t.Fatalf("%q version = %#v, want 1", path, decoded["version"])
			}
		})
	}
}
