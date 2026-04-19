package epic

import (
	"testing"

	"waypoint/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// epicDone locates the epic, collects pending subtasks, shows a confirmation
// (untestable in unit tests), then marks all pending tasks + the epic as done.
// We test the pending-task collection logic and the mark-done mutations.

// =============================================================================
// Pending task collection
// Logic: for each taskID in epic, if taskMap[id] exists and !IsDone → pending
// =============================================================================

func TestEpicDone_CollectPending_AllPending(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	for i := 0; i < 3; i++ {
		task, _ := model.CreateTask("task", []string{"root"})
		_ = epic.AddTaskID(task.GetID())
		taskMap[task.GetID()] = task
	}

	pending := collectPending(epic, taskMap)
	assert.Len(t, pending, 3)
}

func TestEpicDone_CollectPending_NoneRemaining(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	for i := 0; i < 3; i++ {
		task, _ := model.CreateTask("task", []string{"root"})
		task.MarkDone()
		_ = epic.AddTaskID(task.GetID())
		taskMap[task.GetID()] = task
	}

	pending := collectPending(epic, taskMap)
	assert.Empty(t, pending)
}

func TestEpicDone_CollectPending_Mixed(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	doneTask, _ := model.CreateTask("finished", []string{"root"})
	doneTask.MarkDone()
	_ = epic.AddTaskID(doneTask.GetID())
	taskMap[doneTask.GetID()] = doneTask

	pendingTask, _ := model.CreateTask("still-todo", []string{"root"})
	_ = epic.AddTaskID(pendingTask.GetID())
	taskMap[pendingTask.GetID()] = pendingTask

	pending := collectPending(epic, taskMap)
	assert.Len(t, pending, 1)
	assert.Equal(t, "still-todo", pending[0].GetName())
}

func TestEpicDone_CollectPending_NoSubtasks(t *testing.T) {
	epic := newEpic(t, "empty-epic")
	pending := collectPending(epic, map[uuid.UUID]*model.Task{})
	assert.Empty(t, pending)
}

// Orphaned task IDs (deleted task still referenced by epic) are skipped.
func TestEpicDone_CollectPending_OrphanedID_Skipped(t *testing.T) {
	epic := newEpic(t, "release")
	_ = epic.AddTaskID(uuid.New()) // not in taskMap

	pending := collectPending(epic, map[uuid.UUID]*model.Task{})
	assert.Empty(t, pending)
}

// =============================================================================
// Mark all pending tasks done (the mutation epicDone performs)
// =============================================================================

func TestEpicDone_MarkAllPendingDone(t *testing.T) {
	tasks := make([]*model.Task, 3)
	for i := range tasks {
		tasks[i], _ = model.CreateTask("task", []string{"root"})
		assert.False(t, tasks[i].IsDone())
	}

	// simulate the mark-done loop in epicDone
	for _, t := range tasks {
		t.MarkDone()
	}

	for _, task := range tasks {
		assert.True(t, task.IsDone())
	}
}

func TestEpicDone_MarkEpicDone(t *testing.T) {
	epic := newEpic(t, "release")
	assert.False(t, epic.IsDone())
	epic.MarkDone()
	assert.True(t, epic.IsDone())
}

// =============================================================================
// FindEpic interaction (used at start of epicDone)
// =============================================================================

func TestEpicDone_FindEpic_SingleMatch(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "v2-release"),
		newEpic(t, "v3-release"),
	}
	idx, matches := FindEpic("v2-release", epics)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestEpicDone_FindEpic_NoMatch(t *testing.T) {
	epics := []*model.Epic{newEpic(t, "v2-release")}
	idx, matches := FindEpic("nonexistent", epics)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

func TestEpicDone_FindEpic_Ambiguous(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "v2-backend"),
		newEpic(t, "v2-frontend"),
	}
	idx, matches := FindEpic("v2", epics)
	assert.Equal(t, -1, idx)
	assert.Len(t, matches, 2)
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestEpicDoneUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { epicDoneUsage() })
}

// =============================================================================
// Helpers
// =============================================================================

// collectPending replicates the pending-collection loop from epicDone.
func collectPending(epic *model.Epic, taskMap map[uuid.UUID]*model.Task) []*model.Task {
	var pending []*model.Task
	for _, tid := range epic.GetTasksID() {
		if t, ok := taskMap[tid]; ok && !t.IsDone() {
			pending = append(pending, t)
		}
	}
	return pending
}
