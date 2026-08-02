package pipeline

type StepRole string

const (
	RoleRequisite  StepRole = "requisite"
	RoleSource     StepRole = "source"
	RoleTransform  StepRole = "transform"
	RoleAfterwards StepRole = "afterwards"
)

type EffectName string

const EffectCD EffectName = "cd"

type Command struct {
	Argv   []string
	Become bool
}

type Step struct {
	Role    StepRole
	Command Command
	Invert  bool
}

type ResolvedValue struct {
	Name    string
	Value   string
	Trusted bool
}

type Pipeline struct {
	Description string
	Defaults    map[string]string
	Trusted     []string
	Resolved    map[string]ResolvedValue
	Requisite   *Step
	Source      *Step
	Transforms  []Step
	Command     *Command
	Effect      EffectName
	Afterwards  *Step
}
