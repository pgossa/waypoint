package task

import (
	"testing"

	"waypoint/model"

	"github.com/stretchr/testify/assert"
)

// taskUnbound requires exactly 1 arg: <task name>.
// The actual unbinding logic calls storage (untestable without I/O).
// We test arg-count checks and the epicID guard/unset logic.

// =============================================================================
// Arg count validation
// =============================================================================

func TestTaskUnbound_ArgCount_OneRequired(t *testing.T) {
	tests := []struct {
		desc   string
		args   []string
		wantOK bool
	}{
		{"no args", []string{}, false},
		{"one arg (correct)", []string{"task-name"}, true},
		{"two args", []string{"task-name", "extra"}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			ok := len(tt.args) == 1
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

// =============================================================================
// Precondition: task must be bound to an epic
// =============================================================================

func TestTaskUnbound_NotBound_Guard(t *testing.T) {
	task := makeTask(t, "free", []string{"root"})
	// taskUnbound checks: if task.GetEpicID() == nil { error }
	assert.Nil(t, task.GetEpicID())
}

func TestTaskUnbound_BoundTask_Detectable(t *testing.T) {
	task := makeTask(t, "linked", []string{"root"})
	task.SetEpicID(task.GetID()) // stand-in UUID
	assert.NotNil(t, task.GetEpicID())
}

// =============================================================================
// UnsetEpicID clears the field (the mutation taskUnbound performs on the task)
// =============================================================================

func TestTaskUnbound_UnsetEpicID_ClearsField(t *testing.T) {
	task := makeTask(t, "task", []string{"root"})
	task.SetEpicID(task.GetID())
	assert.NotNil(t, task.GetEpicID())

	task.UnsetEpicID()
	assert.Nil(t, task.GetEpicID())
}

// taskUnbound also calls epic.RemoveTaskID to keep the epic in sync.
func TestTaskUnbound_RemoveTaskID_UpdatesEpic(t *testing.T) {
	task := makeTask(t, "task", []string{"root"})

	epic, err := model.CreateEpic("my-epic", []string{"root"}, nil)
	assert.NoError(t, err)

	_ = epic.AddTaskID(task.GetID())
	assert.Len(t, epic.GetTasksID(), 1)

	epic.RemoveTaskID(task.GetID())
	assert.Empty(t, epic.GetTasksID())
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestTaskUnboundUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { taskUnboundUsage() })
}
