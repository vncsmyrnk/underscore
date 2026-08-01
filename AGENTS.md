# Repository instructions

## Product direction

Underscore is a pipeline executor for simple, bounded shell command wrappers.
It discovers strict JSON pipeline contracts from the user's configuration and
executes them by name. It does not discover wrapper implementations from
standalone scripts or executables; executables appear only as argv entries
inside a validated pipeline contract.

The five `docs/workflow-v1*.jsonc` files are the canonical V1 contracts; the
architecture MUST support all of them.

The current Elvish router and working-directory IPC predate this proposal. They
are migration-state code, not architectural precedent, and MUST NOT constrain
the target implementation.

## Product boundary

V1 supports one unified contract with independently bounded roles:

- at most one `requisite`;
- at most one `source`;
- up to four `transform` steps by default;
- at most one terminal `command`;
- at most one `afterwards`; and
- either one `cd` effect or no effect.

The transform limit MAY be overridden explicitly through
`UNDERSCORE_MAX_TRANSFORMS`. The other role limits are fixed in V1.

- **Trigger:** a wrapper needs general branching, retries, loops, concurrency
  beyond the streamed process pipeline, recovery, or always-run cleanup.
  **Action:** reject it as outside the V1 contract.
- **Trigger:** counting commands. **Action:** count every `requisite`, `source`,
  `transform`, terminal `command`, and `afterwards` independently under the
  role limits above. Effects do not count.

## Glossary

- **Pipeline:** the normalized executable contract interpreted by Underscore.
  It may combine lifecycle gates with a streamed process chain and a terminal
  command or shell effect.
- **Lifecycle gate:** an optional non-piped `requisite` before the streamed or
  terminal action, or an optional non-piped `afterwards` command after it.
- **Streamed process chain:** an optional `source`, followed by ordered
  `transform` steps and, when present, a terminal `command`.
- **Wrapper:** a user-facing pipeline discovered by directory name and invoked
  through `underscore` or the `_` shell adapter.
- **Shell adapter:** a shell-specific integration that creates the effect IPC
  channel, invokes `underscore`, validates a structured effect result, and
  applies the effect in the parent shell.

## Discovery and invocation

A pipeline named `<name>` MUST be loaded from:

```text
$XDG_CONFIG_HOME/underscore/pipelines/<name>/config.json
```

When `XDG_CONFIG_HOME` is unset, discovery MUST use
`~/.config/underscore/pipelines/<name>/config.json`.

An optional profile named `<profile-name>` MUST be loaded from:

```text
$XDG_CONFIG_HOME/underscore/pipelines/<name>/profiles/<profile-name>.json
```

The CLI shape is:

```text
underscore <name> [profile-name] [--default-name value ...]
_ <name> [profile-name] [--default-name value ...]
```

The profile name is the only optional positional argument. Pipeline and profile
names MUST match `[A-Za-z0-9][A-Za-z0-9._-]*`. Every CLI flag is a pipeline
default override; framework controls MUST use environment variables, not
reserved flags. Unknown positional arguments, unknown flags, duplicate flags,
unsafe pipeline/profile path segments, and malformed option values MUST be
rejected.

Pipeline and profile configuration paths MAY be symlinks, as required by Nix
and dotfile managers, but their resolved targets MUST be regular files.

Default resolution precedence is:

```text
contract defaults < profile overrides < CLI overrides
```

Profiles MUST be strict JSON objects containing string values only. A profile
or CLI flag MUST NOT introduce a key absent from the contract's `defaults`.

## Shell completion

Zsh completion for both `underscore` and `_` MUST discover pipeline names from
the effective configuration directory at completion time:

```text
$XDG_CONFIG_HOME/underscore/pipelines
```

When `XDG_CONFIG_HOME` is unset, completion MUST use
`~/.config/underscore/pipelines`. It MUST offer only names that satisfy the
pipeline-name grammar and whose directory contains a `config.json` resolving to
a regular file. After a pipeline is selected, completion MUST offer profile
names from that pipeline's `profiles` directory using the same name and
regular-file rules.

- **Trigger:** implementing shell completion. **Action:** use the shell's
  existing completion facilities and filesystem primitives before introducing
  a custom completion generator or runtime dependency.
- **Trigger:** pipelines or profiles are added or removed. **Action:** completion
  MUST reflect the current filesystem without reinstalling or regenerating the
  completion script.

## Manual documentation

The project MUST ship an `underscore(1)` manpage that documents both direct
`underscore` invocation and the `_` shell adapter. Its maintained source MUST
be Markdown, and the build MUST generate the installed manpage with Pandoc
rather than maintaining roff manually.

## Contract format

Runtime contracts MUST be strict, versioned JSON. The canonical `.jsonc` files
are annotated specifications whose behavior MUST be representable as strict
JSON without semantic changes. Unknown fields and unsupported versions MUST be
rejected.

A V1 contract contains:

- `version`: the integer `1`;
- `description`: a required, non-empty prose string;
- `defaults`: an optional object of string values;
- `trusted`: an optional array naming declared defaults;
- `steps`: an optional or null array of step objects;
- `command`: an optional or null non-empty command array;
- `become`: an optional boolean applying to the terminal command; and
- `effect`: an optional supported effect name.

The contract MUST NOT contain a `name`; the discovery directory names the
wrapper.

A command array's first item is the executable; remaining items are ordered
arguments. Every command array MUST be non-empty, and the runtime MUST invoke it
directly without shell parsing.

`command` and `effect` MUST NOT appear together. Omitting `command` and setting
it to null normalize to the same no-command state. Empty strings are invalid.
An effect requires a non-empty streamed process chain. Neither a command nor an
effect is valid only when a non-empty streamed process chain writes its result
to stdout.

`steps` MAY be omitted, empty, or null when a terminal command is present. Each
step has one of four roles:

- `requisite`: a non-piped command that gates all later execution;
- `source`: the first command in a streamed process chain;
- `transform`: an ordered streamed command receiving the previous command's
  stdout on stdin; or
- `afterwards`: a non-piped command that runs only after the action succeeds.

Every step contains `type` and a non-empty `command` array. `become` MAY apply
to any step; `invert` MAY appear only on a requisite. Other step fields are
invalid.

`select` is not a V1 role; interactive filters are `transform` steps.
Transforms require a source. Duplicate singleton roles, unsupported roles,
invalid mixtures, empty pipelines, and contracts over an applicable role limit
MUST be rejected.

Defaults use identifiers of the form `NAME` and are referenced as `$NAME`.
References MAY appear within an argument and MUST expand as data without
changing its argument boundary. The corresponding CLI override is lowercase
kebab case, such as `TARGET_PATH` to `--target-path`. Missing references and
ambiguous flag mappings MUST be rejected.

V1 MUST NOT support command substitution, environment expansion, or a general
template language.

A value listed in `trusted` is an explicit risk annotation for a value passed to
a downstream interpreter, as SOPS interprets `COMMAND`. Every trusted name MUST
identify a declared default. Trust MUST remain present in the normalized model
but MUST NOT change expansion or execution: the runtime passes every resolved
value as one opaque argument and never interprets its content.

## Validation and normalization

Decoded JSON MUST be validated and normalized before execution. Runtime code
MUST consume only normalized values.

Role declaration order MUST normalize to:

```text
requisite -> source -> transforms -> terminal command or effect -> afterwards
```

The relative order of the source and transforms MUST be preserved. If a
terminal command follows a streamed chain, it is the final process in that
chain and receives the previous process's stdout on stdin. Without a streamed
chain, a terminal command inherits stdin.

If no terminal command or effect is present, the final streamed stdout is the
wrapper result. If an effect is present, the effect consumes the final streamed
stdout. Lifecycle gates keep inherited streams and do not join the streamed
chain.

Framework settings are:

- `UNDERSCORE_MAX_TRANSFORMS`: optional non-negative transform limit, default
  `4`;
- `UNDERSCORE_ELEVATION_EXECUTABLE`: one executable name or path, default
  `sudo`; and
- `UNDERSCORE_EFFECT_RESULT`: private effect-result path supplied by a shell
  adapter.

Framework settings MUST NOT be expanded inside contract arguments.

## Execution

Commands MUST run in normalized order and stop later lifecycle phases on
failure. Every process MUST execute with direct argv boundaries.

A `requisite` MUST succeed before the streamed or terminal action starts.
`invert: true` reverses only its ordinary exit-code predicate: zero fails and an
ordinary nonzero status succeeds. Launch errors, signals, and cancellation MUST
NOT be inverted.

The source, transforms, and terminal command MUST run as a concurrent OS
pipeline. The runtime MUST connect each process's stdout to the next process's
stdin, close unused pipe ends promptly, propagate cancellation, and reap every
started child. A partial launch failure MUST cancel and reap already-started
processes.

Streamed execution uses pipefail semantics. If multiple stages fail, the
earliest failing stage in pipeline order determines the reported failure.
SIGPIPE caused by a downstream stage that completed successfully is expected
and MUST NOT fail the wrapper; other signals remain failures.

An `afterwards` command MUST run only after the streamed or terminal action
succeeds. Its failure fails the wrapper. An effect MUST be emitted only after
all scheduled commands, including `afterwards`, succeed.

For a successful pipeline without a terminal command or effect, the final
stream MUST be copied to stdout without whole-output buffering. Terminal command
stdout is inherited. An effect sink MUST drain the final stream while retaining
only the data required to resolve the effect.

Command lifecycle streams are inherited except where the terminal command is
the final consumer of a streamed chain.

## Errors and status

The application MUST use typed errors for discovery, strict decoding,
validation, resolution, process launch, process exit, signal, cancellation,
effect validation, and effect IPC failures.

Every user-facing failure MUST emit one standard stderr message beginning with
`underscore:` and including the relevant pipeline/stage context plus the actual
underlying error description. Requisite failures use the same reporting under
normal and inverted predicates.

On process failure, the failing command's status is the wrapper status. On
success, the terminal command or final streamed stage determines process
status. Non-process failures MUST use documented, stable CLI status classes.
Failed pipelines MUST NOT run later phases or emit an effect.

## Privilege elevation

Top-level `become: true` applies to the terminal command. Step-level
`become: true` applies only to that step's command.

The runtime MUST prefix the affected argv with the configured elevation
executable while preserving the original executable and arguments as separate
arguments. The elevation setting names one executable only and MUST NOT be
shell-parsed. Missing elevation support MUST fail rather than fall back to
unprivileged execution.

## Shell effects and adapters

V1 supports the `cd` effect. The core MUST return a structured shell-neutral
effect; it MUST NOT return arbitrary shell code for `eval` or sourcing.

The effect consumes the first line of the final streamed result and drains the
remaining output. A `cd` result MUST be a non-empty path after removing its
trailing line ending. NUL bytes are invalid.

V1 ships a Zsh adapter through the `_` function. The adapter MUST:

1. create a mode-restricted temporary effect-result file under
   `XDG_RUNTIME_DIR`, falling back safely when it is unavailable;
2. pass its path only through `UNDERSCORE_EFFECT_RESULT`;
3. invoke the `underscore` executable without capturing stdout or stderr;
4. accept only the versioned, NUL-framed effect protocol;
5. validate the complete record and `cd` path before applying
   `builtin cd -- <path>` in the parent shell; and
6. remove the temporary file and restore environment state on every return
   path.

The effect writer MUST reject unsafe result files and MUST NOT follow symlinks.
Running an effect-producing pipeline directly through `underscore` without a
valid adapter channel MUST fail before any pipeline command starts and explain
that `_` is required.

Shell-neutral application and core packages MUST NOT import or encode Zsh
behavior. Future shells MUST be addable as sibling adapters without changing
pipeline policy.

## Technology and delivery

- The interpreter MUST be implemented in Go.
- The bounded variants SHOULD use direct Go structs and focused validation, not
  code generation or a workflow framework.
- The executable MUST discover and execute JSON contracts only.
- The existing command-path router, Elvish runtime dependency, script dispatch,
  and ad hoc working-directory IPC MUST be removed during migration.

```text
name -> strict JSON discovery -> decode -> validate and normalize
-> resolve profile/default overrides -> lifecycle and streamed execution
-> structured result -> optional shell adapter
```

## Go quality gates

- **Trigger:** Go source is added or changed. **Action:** you MUST format it with
  `gofmt` and `goimports`; the formatting gate MUST fail when either tool would
  produce a diff.
- **Trigger:** validating Go changes. **Action:** you MUST run
  `golangci-lint run ./...`, and the gate MUST fail on every enabled finding;
  suppressions MUST be narrow, include a reason, and apply only where the
  finding is intentionally accepted.
- **Trigger:** validating Go tests. **Action:** you MUST run
  `go test -race ./...`; race-detector failures MUST block completion, and
  race-prone code MUST be corrected rather than excluded from the gate.
- **Trigger:** validating Go test coverage. **Action:** you MUST enforce at
  least 90% statement coverage across the module and at least 80% in every
  package; generated code and executable entry points MAY be excluded only
  through an explicit, documented coverage rule.

Nix and the Makefile are the canonical gate entry points. `nix develop` MUST
provide the pinned Go toolchain, `goimports`, `golangci-lint`, Zsh, and test
dependencies. `make check` MUST run formatting checks, lint, race tests,
coverage enforcement, and shell integration tests. `nix flake check`,
`nix develop --command make check`, and `nix build .# -L` MUST remain the CI
contract.

## Go package architecture

Go packages MUST follow an inward dependency direction: `cmd` composes the
executable, adapters implement external integration, application packages
orchestrate use cases, and core packages own domain policy. This keeps policy
independent of delivery mechanisms and operating-system details.

- **Trigger:** adding executable startup code. **Action:** place it under
  `cmd/<name>` and limit it to configuration, dependency construction, signal
  context, and process lifecycle; business and pipeline policy MUST live under
  `internal`.
- **Trigger:** adding domain or pipeline policy. **Action:** place it in a
  focused core package under `internal`; core packages MUST NOT import
  application, adapter, shell, command-entry-point, or infrastructure packages.
- **Trigger:** coordinating a use case. **Action:** place orchestration in an
  application package under `internal`; it MAY import core packages and
  consumer-owned ports, but MUST NOT import concrete adapters.
- **Trigger:** integrating with JSON/filesystem configuration, OS processes,
  shells, effect files, or other external systems. **Action:** place the
  implementation in an adapter package under `internal`; adapters MAY depend
  inward on application or core contracts, MUST NOT define domain policy, and
  MUST NOT import sibling adapters to coordinate work.
- **Trigger:** packages communicate across a boundary. **Action:** use explicit
  function parameters, return values, and narrow interfaces owned by the
  consuming package; MUST NOT use mutable package globals, hidden `init`
  registration, or shared implementation types to bypass the dependency
  direction.
- **Trigger:** a package has no single bounded responsibility or is named
  `common`, `shared`, `helpers`, or `utils`. **Action:** you MUST split or
  relocate it according to the capability that owns the behavior rather than
  creating a generic dependency hub.
- **Trigger:** configuring Go linting. **Action:** you MUST enable `depguard`
  rules that reject imports contrary to the package direction above, imports of
  any `cmd` package, concrete adapter imports outside composition roots, and
  adapter-to-adapter imports; exceptions MUST be explicit, narrow, and
  documented with their architectural reason.

## Canonical V1 contracts

| Contract | Required behavior |
| --- | --- |
| `workflow-v1.jsonc` | `git worktree list` -> interactive `fzf` transform -> directory change using the first output line. |
| `workflow-v1-filter.jsonc` | Package source -> four streamed transforms -> stdout. |
| `workflow-v1-command-check-and-afterwards.jsonc` | Inverted busy requisite -> elevated unmount -> success-only elevated close. |
| `workflow-v1-command.jsonc` | Elevated crypt mapping requisite -> elevated mount. |
| `workflow-v1-no-steps-only.jsonc` | SOPS with configurable secrets path and trusted command text. |

A schema or runtime change that invalidates any canonical contract requires an
explicit architecture decision and a matching contract update.

## Implementation roadmap

The roadmap is contract-first. A wave may start only when all of its declared
dependencies are completed, reviewed, and its produced interfaces are stable.
Tasks in the same parallel group MUST have disjoint primary file ownership. Shared
composition, packaging, and installation files are reserved for integration
tasks to prevent subagent conflicts.

Each task is the smallest independently reviewable unit with its own failing
tests, implementation, and focused checks. Agents MUST use test-driven
development and MUST NOT weaken a gate to accept partial work. Commits remain
subject to explicit user approval and, when requested, MUST use Conventional
Commits.

### Wave 0: Nix and Makefile foundation

#### Task W0-A: Establish the Go and quality-gate baseline

**Owner:** one foundation subagent; this task is sequential and blocks all Go
implementation.

**Files:**

- Create `go.mod`.
- Create `.golangci.yml`.
- Create `internal/testsupport/commandtest/main.go` for deterministic child
  process scenarios used by later tests.
- Create `internal/testsupport/coverage/coverage.go` and tests for module and
  per-package threshold enforcement.
- Modify `Makefile`.
- Modify `flake.nix`; modify `flake.lock` only if an input revision is
  deliberately changed.
- Modify `.github/workflows/ci.yaml` only if the existing three-command CI
  contract cannot invoke the new gates unchanged.

**Deliverable:** `make fmt`, `make fmt-check`, `make lint`, `make test`,
`make coverage`, `make check`, and `make build` use tools supplied by the pinned
Nix flake. `make test` runs `go test -race ./...`; coverage enforces 90% module
and 80% per package; formatting checks both `gofmt` and `goimports`;
`golangci-lint` enables architectural `depguard` rules. Coverage exclusions are
documented and limited to `cmd` entry points and the deterministic child-process
test executable.

**Gate:**

```text
nix flake check
nix develop --command make check
nix build .# -L
```

### Wave 1: Interface freeze

Wave 1 depends on W0-A and is owned by one contract subagent because every
parallel implementation track consumes these definitions.

#### Task W1-A: Define normalized pipeline and application ports

**Files:**

- Create `internal/core/pipeline/types.go`.
- Create `internal/core/pipeline/limits.go`.
- Create `internal/core/pipeline/errors.go`.
- Create `internal/core/pipeline/types_test.go`.
- Create `internal/application/run/ports.go`.
- Create `internal/application/run/result.go`.
- Create `internal/application/run/ports_test.go`.
- Create `internal/adapter/effectfile/protocol.go`.
- Create `internal/adapter/effectfile/protocol_test.go`.
- Create strict JSON equivalents of all canonical contracts under
  `internal/core/pipeline/testdata/`.

**Produced interfaces:** immutable normalized definitions for commands, role
steps, resolved arguments, trust annotations, effects, plans, process outcomes,
typed errors, application results, configuration loading, process execution,
and effect writing. The process port MUST return one outcome per pipeline stage
so application policy can implement ordered pipefail without importing OS
process types. The effect protocol MUST freeze its magic, version, record
fields, and maximum record size.

**Acceptance:** compile-time dependency tests demonstrate the inward package
direction, every canonical strict fixture is valid JSON with an explicit
expected normalized model for W2-A, and protocol round-trip tests cover
truncation, extra fields, unknown versions, and trailing bytes.

### Wave 2: Parallel capability tracks

All Wave 2 tasks depend on W1-A. They form one parallel group and MUST be
assigned to separate subagents.

#### Task W2-A: Strict decode, validation, normalization, and expansion

**Primary ownership:** `internal/core/pipeline/**`, excluding Wave 1 files except
for additive interface-compatible changes approved by the integrator.

**Files:**

- Create `internal/core/pipeline/decode.go` and `decode_test.go`.
- Create `internal/core/pipeline/validate.go` and `validate_test.go`.
- Create `internal/core/pipeline/normalize.go` and `normalize_test.go`.
- Create `internal/core/pipeline/expand.go` and `expand_test.go`.

**Deliverable:** strict `version: 1` JSON decoding; unknown-field rejection;
every role/cardinality/terminal/default/trust/reference rule; normalized
lifecycle and stream ordering; opaque in-argument expansion; deterministic
uppercase-underscore to lowercase-kebab flag mapping; and typed errors with
field/stage context.

**Focused gate:** table-driven tests cover all valid canonical shapes plus every
rejected mixture, duplicate role, limit violation, missing reference, invalid
command, invalid trust name, and ambiguous flag mapping.

#### Task W2-B: Filesystem discovery, profiles, and invocation parsing

**Primary ownership:** `internal/adapter/configfs/**` and
`internal/adapter/cli/**`.

**Files:**

- Create `internal/adapter/configfs/store.go` and `store_test.go`.
- Create `internal/adapter/configfs/profile.go` and `profile_test.go`.
- Create `internal/adapter/cli/invocation.go` and `invocation_test.go`.
- Create `internal/adapter/cli/overrides.go` and `overrides_test.go`.

**Deliverable:** XDG-aware contract/profile lookup; safe single-segment names;
regular-file checks; strict string-only profile objects; two-stage parsing of
`<name> [profile] [flags]`; duplicate/unknown option rejection; and precedence
assembly of defaults, profile, and CLI values without importing concrete
process or shell adapters.

**Focused gate:** tests use isolated temporary config homes and cover unset
XDG state, missing files, traversal attempts, symlinks to regular files,
symlinks to non-files, malformed profiles, unknown keys, duplicate flags,
kebab mapping, and the exact precedence order.

#### Task W2-C: OS process, pipeline, and elevation adapter

**Primary ownership:** `internal/adapter/process/**`.

**Files:**

- Create `internal/adapter/process/runner.go` and `runner_test.go`.
- Create `internal/adapter/process/pipeline.go` and `pipeline_test.go`.
- Create `internal/adapter/process/elevation.go` and `elevation_test.go`.
- Create `internal/adapter/process/outcome.go`.

**Deliverable:** direct argv launch, inherited lifecycle streams, concurrent
pipeline startup, pipe closure, full child reaping, cancellation and signal
propagation, ordered outcomes, partial-launch cleanup, and one-executable
elevation prefixing. The adapter reports facts; it MUST NOT decide requisite
inversion, ordered pipefail, or afterwards policy.

**Focused gate:** helper-process tests prove argv boundaries, stdin/stdout
streaming and backpressure, no whole-output buffering, downstream-success
SIGPIPE classification, multiple failures, cancellation, launch failure after
partial startup, elevation argv, and absence of leaked children or descriptors.

#### Task W2-D: Application lifecycle and pipeline orchestration

**Primary ownership:** `internal/application/run/**`, excluding Wave 1 files
except for interface-compatible additions approved by the integrator.

**Files:**

- Create `internal/application/run/service.go` and `service_test.go`.
- Create `internal/application/run/requisite.go` and `requisite_test.go`.
- Create `internal/application/run/action.go` and `action_test.go`.
- Create `internal/application/run/afterwards.go` and `afterwards_test.go`.
- Create `internal/application/run/status.go` and `status_test.go`.

**Deliverable:** orchestration of requisite, streamed/terminal action, and
afterwards; ordinary-exit-only inversion; ordered pipefail; expected SIGPIPE
handling; stdout streaming; effect draining and first-line extraction; no later
phase after failure; no effect before all commands succeed; and stable typed
status classes.

**Focused gate:** use only fake consumer-owned ports. Tests cover every phase
transition and failure edge, including inverted zero/nonzero, launch/signal
non-inversion, multiple pipeline outcomes, afterwards suppression/failure,
empty/NUL effect targets, failed pipelines returning no effect, and exact
stderr-ready error context.

#### Task W2-E: Effect IPC and pluggable Zsh shell adapter

**Primary ownership:** `internal/adapter/effectfile/**` beyond protocol files,
`shell/entrypoint/zsh`, and new Zsh tests under `shell/entrypoint/`.

**Files:**

- Create `internal/adapter/effectfile/writer.go` and `writer_test.go`.
- Create `internal/adapter/effectfile/security_unix.go` and
  `security_unix_test.go`.
- Replace `shell/entrypoint/zsh`.
- Create `shell/entrypoint/zsh_test.zsh`.

**Deliverable:** safe, bounded, atomic effect records; regular-file,
ownership/mode, and no-symlink checks; a Zsh `_` function that preserves
streaming stdout/stderr and status; complete NUL-framed protocol validation;
parent-shell `builtin cd`; trap-safe cleanup; and restoration of a pre-existing
`UNDERSCORE_EFFECT_RESULT` value. The implementation MUST keep all Zsh behavior
outside core and application packages.

**Focused gate:** `zsh -f` tests cover no-effect execution, successful directory
change, paths containing spaces or glob characters,
malformed/truncated/oversized records, unknown versions/effects, nonzero
executable status, missing binary, cleanup, and environment restoration.

### Wave 3: Integration and migration

Wave 3 starts after all Wave 2 tasks pass focused review. Shared-file tasks are
serial.

#### Task W3-A: Compose the `underscore` executable

**Owner:** one integration subagent.

**Files:**

- Create `cmd/underscore/main.go` and `main_test.go`.
- Create `internal/adapter/cli/errors.go` and `errors_test.go`.
- Create `internal/application/run/load.go` and `load_test.go`.

**Deliverable:** signal-aware composition of config, contract, process, and
effect adapters; two-stage invocation parsing; framework environment parsing;
stable `underscore:` diagnostics; preservation of process exit status;
documented non-process status classes; and pre-execution rejection of an effect
contract without a valid adapter channel.

**Acceptance:** black-box CLI tests run against isolated config homes and prove
that pipeline names—not internal contract fields or script paths—select work.

#### Task W3-B: Replace Elvish delivery and installation surfaces

**Owner:** one migration subagent after W3-A.

**Files:**

- Modify `Makefile`.
- Modify `flake.nix`; modify `flake.lock` only if an input revision is
  deliberately changed.
- Modify `README.md`.
- Modify `shell/rc/.zshrc` if it remains necessary.
- Delete the root Elvish `underscore` script.
- Delete `scripts/init.elv`.
- Delete `scripts/git/worktree/cd.elv`.
- Delete `t/underscore.t`.

**Deliverable:** build and install the Go binary plus the pluggable Zsh adapter
while removing Elvish, script dispatch, nested command-path completion, and ad
hoc cwd IPC dependencies; document config layout, strict JSON contracts,
profiles, dynamic flags, environment controls, direct invocation, and `_`
effect invocation.

**Acceptance:** installation into a temporary prefix contains the intended
binary and Zsh adapter; uninstall removes exactly those artifacts; the README
describes shipped behavior rather than migration state.

### Wave 4: Canonical acceptance

Wave 4 depends on W3-B. Its test-only tasks may run in parallel.

#### Task W4-A: Canonical contract acceptance

**Primary ownership:** `internal/acceptance/**`.

**Files:**

- Create `internal/acceptance/canonical_test.go`.
- Create `internal/acceptance/failure_test.go`.
- Create `internal/acceptance/testdata/` fixtures mirroring all five canonical
  contracts with deterministic helper commands.

**Deliverable:** end-to-end coverage of worktree selection to `cd`, four
streamed transforms to stdout, elevated mount after a requisite, inverted busy
check plus elevated unmount and success-only close, and SOPS-style trusted
opaque command text. Failure cases prove underlying error text, exact status,
no later command, and no effect.

#### Task W4-B: Streaming, security, and shell acceptance

**Primary ownership:** `internal/acceptance/**` files disjoint from W4-A and
`shell/entrypoint/acceptance_test.zsh`.

**Files:**

- Create `internal/acceptance/streaming_test.go`.
- Create `internal/acceptance/security_test.go`.
- Create `shell/entrypoint/acceptance_test.zsh`.

**Deliverable:** integration proof for large streaming data and backpressure,
ordered pipefail, accepted versus real SIGPIPE, cancellation/reaping, traversal
rejection, configuration symlinks to regular files, unsafe effect-result
symlink rejection, opaque argument boundaries, secure effect files, direct
effect refusal, and actual parent-Zsh directory mutation through `_`.

### Wave 5: Completion, manual documentation, and release

Wave 5 starts only after W4-A and W4-B pass. Completion and manpage authoring
are equal-priority sibling tasks with disjoint ownership and MAY run in
parallel. Their shared build, packaging, and installation surfaces are reserved
for W5-C. This keeps both capabilities late in delivery without allowing either
to delay core implementation or acceptance work.

#### Task W5-A: Generate dynamic Zsh completion

**Primary ownership:** `completions/zsh` and completion-specific Zsh tests under
`shell/entrypoint/`.

**Files:**

- Replace `completions/zsh`.
- Create `shell/entrypoint/completion_test.zsh`.

**Deliverable:** completion for both `underscore` and `_` uses native Zsh
completion facilities to inspect the effective pipelines directory on every
completion request. It offers only valid, existing pipeline directories whose
`config.json` resolves to a regular file, then valid profiles whose JSON file
resolves to a regular file. It MUST NOT embed pipeline names, require
regeneration after configuration changes, execute a pipeline, or add a custom
completion framework.

**Focused gate:** `zsh -f` tests cover explicit and unset `XDG_CONFIG_HOME`,
pipeline and profile additions and removals without reinstalling completion,
invalid names, missing configurations, directories in place of JSON files,
symlinks to regular files, symlinks to non-files, spaces in the configuration
home, and identical candidates for `underscore` and `_`.

#### Task W5-B: Author and generate the manpage

**Primary ownership:** manpage source and Pandoc availability in `flake.nix`.

**Files:**

- Create `docs/underscore.1.md`.
- Modify `flake.nix` to provide Pandoc without changing input revisions.

**Deliverable:** one Markdown source describes synopsis, discovery paths,
contract and profile resolution, dynamic overrides, framework environment,
exit behavior, direct invocation, `_` effect invocation, files, environment,
diagnostics, and examples. Pandoc generates a valid `underscore(1)` document;
generated roff MUST NOT be maintained as a second source.

**Focused gate:** generate the manpage with
`pandoc --standalone --to man --output dist/underscore.1
docs/underscore.1.md`, render it with the available manpage viewer, and check
that required sections and both invocation forms are present.

#### Task W5-C: Integrate completion and manpage delivery

**Owner:** one delivery integration subagent after W5-A and W5-B.

**Files:**

- Modify `Makefile`.
- Modify `flake.nix` only for package-source or installation wiring not owned by
  W5-B.
- Modify `README.md`.

**Deliverable:** canonical build targets generate `dist/underscore.1` with
Pandoc, install the Zsh completion and manpage under the prefix's standard
`share/zsh/site-functions` and `share/man/man1` locations, and remove those
artifacts on uninstall. The README documents completion discovery and
`man underscore` without duplicating the manpage.

**Acceptance:** installation into a temporary prefix contains the binary, Zsh
adapter, dynamic completion, and generated manpage in their intended locations;
completion observes pipelines created after installation; `man` renders the
installed page; uninstall removes exactly the installed artifacts.

#### Task W5-D: Run the release gate

**Owner:** one release subagent after W5-C.

**Files:** no feature files; only narrow fixes in the owning task's files may be
made after routing failures back to that owner.

**Gate:**

```text
nix flake check
nix develop --command make fmt-check
nix develop --command make lint
nix develop --command make test
nix develop --command make coverage
nix develop --command make check
nix build .# -L
```

The release review MUST also confirm that no Go package imports against the
declared dependency direction, no Elvish/script router or cwd IPC remains, every
canonical behavior maps to an acceptance test, and the module/package coverage
thresholds are met without broad exclusions.

### Parallel dispatch map

```text
W0-A
  -> W1-A
     -> [W2-A, W2-B, W2-C, W2-D, W2-E] in parallel
        -> W3-A
           -> W3-B
              -> [W4-A, W4-B] in parallel
                 -> [W5-A, W5-B] in parallel
                    -> W5-C
                       -> W5-D
```

The Wave 2 subagents MUST receive the frozen W1-A interfaces and their primary
ownership list in their prompts. They MUST NOT edit another track's files.
Cross-track interface pressure is reported to the integration owner, who either
updates W1-A once for all consumers or keeps the original contract; subagents
MUST NOT make unilateral shared-interface changes.

## Completion criteria

V1 is complete when all canonical specifications are expressible as strict JSON
and execute by discovered name; profile and CLI overrides resolve with the
defined precedence; lifecycle gates and streamed stages obey their independent
bounds; terminal commands consume pipeline stdin; requisite inversion,
streaming, pipefail, expected SIGPIPE, elevation, status propagation, and trust
behave as specified; failures include their underlying cause, do not continue
to later phases, and return no effect; shell mutation is limited to validated
structured effects through the pluggable Zsh adapter; and the Elvish router,
script dispatch, and ad hoc working-directory IPC are absent. Zsh completion
discovers existing pipelines and profiles from the effective configuration
directory without regeneration, and the installed `underscore(1)` manpage is
generated from its Markdown source with Pandoc.
