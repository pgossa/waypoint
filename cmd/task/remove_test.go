package task

import (
	"testing"

	"waypoint/model"

	"github.com/stretchr/testify/assert"
)

// taskRemove calls findTask to locate the target, then runs a huh confirmation
// (untestable in unit tests). We test the findTask logic and removal preconditions.

// =============================================================================
// findTask resolution (as used by taskRemove)
// =============================================================================

func TestTaskRemove_NoMatchFound(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "build-api", []string{"root"}),
		makeTask(t, "fix-tests", []string{"root"}),
	}
	idx, matches := findTask("nonexistent", tasks)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

func TestTaskRemove_SingleMatchByName(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "build-api", []string{"root"}),
		makeTask(t, "fix-tests", []string{"root"}),
	}
	idx, matches := findTask("build-api", tasks)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestTaskRemove_SingleMatchByPartialName(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "build-api", []string{"root"}),
		makeTask(t, "fix-tests", []string{"root"}),
	}
	idx, matches := findTask("fix", tasks)
	assert.Equal(t, 1, idx)
	assert.Equal(t, []int{1}, matches)
}

func TestTaskRemove_MultipleMatches(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "build-api", []string{"root"}),
		makeTask(t, "build-worker", []string{"root"}),
		makeTask(t, "fix-tests", []string{"root"}),
	}
	idx, matches := findTask("build", tasks)
	assert.Equal(t, -1, idx)
	assert.ElementsMatch(t, []int{0, 1}, matches)
}

// =============================================================================
// Task properties used by remove
// =============================================================================

func TestTaskRemove_TaskHasStableID(t *testing.T) {
	task := makeTask(t, "removable", []string{"root"})
	assert.Equal(t, task.GetID(), task.GetID())
	assert.NotEmpty(t, task.GetID())
}

func TestTaskRemove_DifferentTasks_DifferentIDs(t *testing.T) {
	a := makeTask(t, "task-a", []string{"root"})
	b := makeTask(t, "task-b", []string{"root"})
	assert.NotEqual(t, a.GetID(), b.GetID())
}

// taskRemove removes both pending and done tasks (no IsDone guard unlike taskDone).
func TestTaskRemove_CanRemovePendingOrDone(t *testing.T) {
	pending := makeTask(t, "pending", []string{"root"})
	done := makeTask(t, "done-task", []string{"root"})
	done.MarkDone()

	assert.NotEmpty(t, pending.GetID())
	assert.NotEmpty(t, done.GetID())
}

// =============================================================================
// Epic linkage: taskRemove must also update the parent epic
// We verify the predicate that drives the epic-update branch.
// =============================================================================

func TestTaskRemove_EpicBoundTask_HasEpicID(t *testing.T) {
	epicTask, err := model.CreateTask("linked", []string{"root"})
	assert.NoError(t, err)

	epicID := epicTask.GetID() // use task ID as a stand-in UUID
	epicTask.SetEpicID(epicID)

	assert.NotNil(t, epicTask.GetEpicID(), "task bound to epic should have non-nil epicID")
}

func TestTaskRemove_UnboundTask_HasNoEpicID(t *testing.T) {
	task := makeTask(t, "standalone", []string{"root"})
	assert.Nil(t, task.GetEpicID())
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestTaskRemoveUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { taskRemoveUsage() })
}
