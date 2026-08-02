package pipeline

import "strings"

type ErrorKind string

const (
	ErrDiscovery        ErrorKind = "discovery"
	ErrDecode           ErrorKind = "decode"
	ErrValidation       ErrorKind = "validation"
	ErrResolution       ErrorKind = "resolution"
	ErrLaunch           ErrorKind = "launch"
	ErrExit             ErrorKind = "exit"
	ErrSignal           ErrorKind = "signal"
	ErrCancellation     ErrorKind = "cancellation"
	ErrEffectValidation ErrorKind = "effect_validation"
	ErrEffectIPC        ErrorKind = "effect_ipc"
)

type Error struct {
	Kind     ErrorKind
	Pipeline string
	Field    string
	Stage    string
	Err      error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	parts := []string{string(e.Kind)}
	if e.Pipeline != "" {
		parts = append(parts, "pipeline="+e.Pipeline)
	}
	if e.Field != "" {
		parts = append(parts, "field="+e.Field)
	}
	if e.Stage != "" {
		parts = append(parts, "stage="+e.Stage)
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}

	return strings.Join(parts, ": ")
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}
