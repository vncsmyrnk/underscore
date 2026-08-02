package effectfile

import (
	"fmt"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

const (
	ProtocolMagic          = "underscore-effect"
	ProtocolVersion uint16 = 1
	MaxRecordSize          = 4096
)

type Record struct {
	Magic   string              `json:"magic"`
	Version uint16              `json:"version"`
	Effect  pipeline.EffectName `json:"effect"`
	Value   string              `json:"value"`
}

func ValidateRecord(record Record) error {
	switch {
	case record.Magic != ProtocolMagic:
		return fmt.Errorf("invalid protocol magic %q", record.Magic)
	case record.Version != ProtocolVersion:
		return fmt.Errorf("invalid protocol version %d", record.Version)
	case record.Effect == "":
		return fmt.Errorf("effect is required")
	case record.Value == "":
		return fmt.Errorf("value is required")
	case len(record.Value) > MaxRecordSize:
		return fmt.Errorf("record value exceeds %d bytes", MaxRecordSize)
	default:
		return nil
	}
}
