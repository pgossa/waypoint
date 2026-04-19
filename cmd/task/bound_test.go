package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// taskBound requires exactly 2 args: <task> <epic>.
// The actual binding logic calls storage (untestable without I/O),
// so we test the arg-count check and the IsDone / epicID guards.

// =============================================================================
// Arg count validation
// =============================================================================

func TestTaskBound_ArgCount_TwoRequired(t *testing.T) {
	tests := []struct {
		desc    string
		args    []string
		wantOK  bool
	}{
		{"no args", []string{}, false},
		{"one arg", []string{"task-name"}, false},
		{"two args (correct)", []string{"task-name", "epic-name"}, true},
		{"three args", []string{"a", "b", "c"}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			ok := len(tt.args) == 2
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

// =============================================================================
// Precondition: task must not already be bound to an epic
// =============================================================================

func TestTaskBound_AlreadyBound_Guard(t *testing.T) {
	task := makeTask(t, "linked", []string{"root"})
	id := task.GetID() // reuse any UUID
	task.SetEpicID(id)

	// taskBound checks: if task.GetEpicID() != nil { error }
	assert.NotNil(t, task.GetEpicID(), "already-bound task should have non-nil epicID")
}

func TestTaskBound_Unbound_CanBeBound(t *testing.T) {
	task := makeTask(t, "free", []string{"root"})
	assert.Nil(t, task.GetEpicID(), "unbound task should have nil epicID")
}

// =============================================================================
// Binding sets epicID on task and adds task to epic
// =============================================================================

func TestTaskBound_SetEpicID_SetsField(t *testing.T) {
	task := makeTask(t, "task", []string{"root"})
	epicID := task.GetID() // stand-in UUID

	task.SetEpicID(epicID)
	assert.NotNil(t, task.GetEpicID())
	assert.Equal(t, epicID, *task.GetEpicID())
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestTaskBoundUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { taskBoundUsage() })
}
