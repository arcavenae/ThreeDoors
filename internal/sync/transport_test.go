package sync_test

import (
	"testing"

	gosync "github.com/arcaven/ThreeDoors/internal/sync"
)

func TestSyncOp_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		op   gosync.SyncOp
		want string
	}{
		{"add", gosync.OpAdd, "add"},
		{"modify", gosync.OpModify, "modify"},
		{"delete", gosync.OpDelete, "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.op.String(); got != tt.want {
				t.Errorf("SyncOp.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChangeset_Empty(t *testing.T) {
	t.Parallel()

	cs := gosync.Changeset{}
	if len(cs.Files) != 0 {
		t.Errorf("empty Changeset should have 0 files, got %d", len(cs.Files))
	}
}
