package task

import (
	"testing"

	"waypoint/model"

	"github.com/stretchr/testify/assert"
)

// taskDone calls findTask then checks IsDone before marking.
// findTask is already tested in common_test.go; here we test the
// IsDone guard and the MarkDone transition that taskDone relies on.

// =============================================================================
// IsDone guard: taskDone exits early if the task is already done
// =============================================================================

func TestTaskDone_AlreadyDone_Guard(t *testing.T) {
	task := makeTask(t, "completed", []string{"root"})
	task.MarkDone()
	// taskDone would call ui.Error + os.Exit here.
	// We verify the IsDone predicate that drives that branch.
	assert.True(t, task.IsDone(), "task should be done before re-attempting mark")
}

func TestTaskDone_MarkDone_Transition(t *testing.T) {
	task := makeTask(t, "pending", []string{"root"})
	assert.False(t, task.IsDone())
	task.MarkDone()
	assert.True(t, task.IsDone())
}

func TestTaskDone_MarkDone_Idempotent(t *testing.T) {
	task := makeTask(t, "idempotent", []string{"root"})
	task.MarkDone()
	task.MarkDone()
	assert.True(t, task.IsDone())
}

// =============================================================================
// findTask resolution (as used inside taskDone)
// =============================================================================

func TestTaskDone_FindTask_SingleMatch(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "write-tests", []string{"root"}),
		makeTask(t, "fix-bug", []string{"root"}),
	}
	idx, matches := findTask("write-tests", tasks)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestTaskDone_FindTask_NoMatch(t *testing.T) {
	tasks := []*model.Task{makeTask(t, "write-tests", []string{"root"})}
	idx, matches := findTask("nonexistent", tasks)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

func TestTaskDone_FindTask_Ambiguous(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "write-tests", []string{"root"}),
		makeTask(t, "write-docs", []string{"root"}),
	}
	idx, matches := findTask("write", tasks)
	assert.Equal(t, -1, idx)
	assert.Len(t, matches, 2)
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestTaskDoneUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { taskDoneUsage() })
}
