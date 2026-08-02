package effectfile

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/vncsmyrnk/underscore/internal/core/pipeline"
)

func TestRecordRoundTripsThroughJSON(t *testing.T) {
	t.Parallel()

	record := Record{
		Magic:   ProtocolMagic,
		Version: ProtocolVersion,
		Effect:  pipeline.EffectCD,
		Value:   "/tmp/worktree",
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	decoded, err := decodeRecord(data)
	if err != nil {
		t.Fatalf("decodeRecord returned error: %v", err)
	}

	if decoded != record {
		t.Fatalf("decoded record = %#v, want %#v", decoded, record)
	}
}

func TestValidateRecordRejectsInvalidData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		record Record
	}{
		{
			name: "wrong magic",
			record: Record{
				Magic:   "wrong",
				Version: ProtocolVersion,
				Effect:  pipeline.EffectCD,
				Value:   "/tmp/worktree",
			},
		},
		{
			name: "wrong version",
			record: Record{
				Magic:   ProtocolMagic,
				Version: ProtocolVersion + 1,
				Effect:  pipeline.EffectCD,
				Value:   "/tmp/worktree",
			},
		},
		{
			name: "missing effect",
			record: Record{
				Magic:   ProtocolMagic,
				Version: ProtocolVersion,
				Value:   "/tmp/worktree",
			},
		},
		{
			name: "missing value",
			record: Record{
				Magic:   ProtocolMagic,
				Version: ProtocolVersion,
				Effect:  pipeline.EffectCD,
			},
		},
		{
			name: "oversized value",
			record: Record{
				Magic:   ProtocolMagic,
				Version: ProtocolVersion,
				Effect:  pipeline.EffectCD,
				Value:   strings.Repeat("a", MaxRecordSize+1),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := ValidateRecord(tt.record); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateRecordAcceptsValidRecord(t *testing.T) {
	t.Parallel()

	record := Record{
		Magic:   ProtocolMagic,
		Version: ProtocolVersion,
		Effect:  pipeline.EffectCD,
		Value:   "/tmp/worktree",
	}

	if err := ValidateRecord(record); err != nil {
		t.Fatalf("ValidateRecord returned error: %v", err)
	}
}

func TestDecodeRecordRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "unknown field",
			data: []byte(`{"magic":"underscore-effect","version":1,"effect":"cd","value":"/tmp/worktree","extra":"nope"}`),
		},
		{
			name: "truncated",
			data: []byte(`{"magic":"underscore-effect","version":1`),
		},
		{
			name: "trailing bytes",
			data: []byte(`{"magic":"underscore-effect","version":1,"effect":"cd","value":"/tmp/worktree"}oops`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := decodeRecord(tt.data); err == nil {
				t.Fatal("expected decode error")
			}
		})
	}
}

func decodeRecord(data []byte) (Record, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var record Record
	if err := decoder.Decode(&record); err != nil {
		return Record{}, err
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Record{}, io.ErrUnexpectedEOF
		}

		return Record{}, err
	}

	return record, nil
}
