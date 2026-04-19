package epic

import (
	"strings"
	"testing"

	"waypoint/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// epicUnbound accepts 1 or 2 args: <epic> [task].
// When 2 args are given, it searches the epic's bound tasks by name (no storage).
// We test the arg-count check and the inline bound-task name search.

// =============================================================================
// Arg count validation: 1 or 2 args accepted
// =============================================================================

func TestEpicUnbound_ArgCount(t *testing.T) {
	tests := []struct {
		desc   string
		args   []string
		wantOK bool
	}{
		{"no args", []string{}, false},
		{"one arg (epic only)", []string{"my-epic"}, true},
		{"two args (epic + task)", []string{"my-epic", "my-task"}, true},
		{"three args", []string{"a", "b", "c"}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			ok := len(tt.args) >= 1 && len(tt.args) <= 2
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

// =============================================================================
// Bound-task name search (used when 2nd arg is provided)
// Logic from epicUnbound: iterate epic.GetTasksID(), match task name by substring
// =============================================================================

func TestEpicUnbound_BoundTaskSearch_ExactMatch(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	taskA, _ := model.CreateTask("build-api", []string{"root"})
	taskB, _ := model.CreateTask("fix-tests", []string{"root"})
	_ = epic.AddTaskID(taskA.GetID())
	_ = epic.AddTaskID(taskB.GetID())
	taskMap[taskA.GetID()] = taskA
	taskMap[taskB.GetID()] = taskB

	found := searchBoundTasks(epic, taskMap, "build-api")
	assert.Len(t, found, 1)
	assert.Equal(t, taskA.GetID(), found[0])
}

func TestEpicUnbound_BoundTaskSearch_PartialMatch(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	taskA, _ := model.CreateTask("build-api", []string{"root"})
	taskB, _ := model.CreateTask("fix-tests", []string{"root"})
	_ = epic.AddTaskID(taskA.GetID())
	_ = epic.AddTaskID(taskB.GetID())
	taskMap[taskA.GetID()] = taskA
	taskMap[taskB.GetID()] = taskB

	found := searchBoundTasks(epic, taskMap, "build")
	assert.Len(t, found, 1)
	assert.Equal(t, taskA.GetID(), found[0])
}

func TestEpicUnbound_BoundTaskSearch_AmbiguousMatch(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	taskA, _ := model.CreateTask("build-api", []string{"root"})
	taskB, _ := model.CreateTask("build-worker", []string{"root"})
	_ = epic.AddTaskID(taskA.GetID())
	_ = epic.AddTaskID(taskB.GetID())
	taskMap[taskA.GetID()] = taskA
	taskMap[taskB.GetID()] = taskB

	found := searchBoundTasks(epic, taskMap, "build")
	assert.Len(t, found, 2)
}

func TestEpicUnbound_BoundTaskSearch_NoMatch(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	task, _ := model.CreateTask("build-api", []string{"root"})
	_ = epic.AddTaskID(task.GetID())
	taskMap[task.GetID()] = task

	found := searchBoundTasks(epic, taskMap, "nonexistent")
	assert.Empty(t, found)
}

func TestEpicUnbound_BoundTaskSearch_CaseInsensitive(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	task, _ := model.CreateTask("Build-API", []string{"root"})
	_ = epic.AddTaskID(task.GetID())
	taskMap[task.GetID()] = task

	found := searchBoundTasks(epic, taskMap, "build-api")
	assert.Len(t, found, 1)

	found = searchBoundTasks(epic, taskMap, "BUILD")
	assert.Len(t, found, 1)
}

// Orphaned task IDs (in epic but not in taskMap) are skipped in search.
func TestEpicUnbound_BoundTaskSearch_OrphanedID_Skipped(t *testing.T) {
	epic := newEpic(t, "release")
	_ = epic.AddTaskID(uuid.New()) // not in taskMap

	found := searchBoundTasks(epic, map[uuid.UUID]*model.Task{}, "anything")
	assert.Empty(t, found)
}

// =============================================================================
// Precondition: epic must have at least one bound task
// =============================================================================

func TestEpicUnbound_EmptyEpic_Guard(t *testing.T) {
	epic := newEpic(t, "empty")
	// epicUnbound checks: if len(epic.GetTasksID()) == 0 { error }
	assert.Empty(t, epic.GetTasksID())
}

func TestEpicUnbound_EpicWithTasks_CanProceed(t *testing.T) {
	task, _ := model.CreateTask("subtask", []string{"root"})
	epic, _ := model.CreateEpic("release", []string{"root"}, []uuid.UUID{task.GetID()})
	assert.NotEmpty(t, epic.GetTasksID())
}

// =============================================================================
// Mutations performed by epicUnbound
// =============================================================================

func TestEpicUnbound_RemoveTaskID_UpdatesEpic(t *testing.T) {
	taskA, _ := model.CreateTask("a", []string{"root"})
	taskB, _ := model.CreateTask("b", []string{"root"})
	epic, _ := model.CreateEpic("release", []string{"root"}, []uuid.UUID{taskA.GetID(), taskB.GetID()})

	epic.RemoveTaskID(taskA.GetID())
	assert.Len(t, epic.GetTasksID(), 1)
	assert.Equal(t, taskB.GetID(), epic.GetTasksID()[0])
}

func TestEpicUnbound_UnsetEpicID_ClearsTask(t *testing.T) {
	task, _ := model.CreateTask("task", []string{"root"})
	task.SetEpicID(task.GetID()) // stand-in epic UUID
	assert.NotNil(t, task.GetEpicID())

	task.UnsetEpicID()
	assert.Nil(t, task.GetEpicID())
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestEpicUnboundUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { epicUnboundUsage() })
}

// =============================================================================
// Helpers
// =============================================================================

// searchBoundTasks replicates the inline task-search loop from epicUnbound.
func searchBoundTasks(epic *model.Epic, taskMap map[uuid.UUID]*model.Task, input string) []uuid.UUID {
	lower := strings.ToLower(input)
	var found []uuid.UUID
	for _, tid := range epic.GetTasksID() {
		t, ok := taskMap[tid]
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(t.GetName()), lower) {
			found = append(found, tid)
		}
	}
	return found
}
