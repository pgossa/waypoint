package storage

import (
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Tests for the public storage dispatcher (storage.go).
// Uses setupTempStorage from json_test.go — sets HOME to a temp dir so the
// JSON backend writes to an isolated location, and resets config cache.

// =============================================================================
// Job — SaveJob / GetJobs / RemoveJob
// =============================================================================

func TestDispatcher_SaveAndGetJob(t *testing.T) {
	setupTempStorage(t)

	job, _ := model.CreateJob("buy milk")
	assert.NoError(t, SaveJob(job))

	jobs, err := GetJobs()
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, job.GetID(), jobs[0].GetID())
	assert.Equal(t, "buy milk", jobs[0].GetName())
	assert.False(t, jobs[0].IsDone())
}

func TestDispatcher_SaveJob_UpdatesExisting(t *testing.T) {
	setupTempStorage(t)

	job, _ := model.CreateJob("pending")
	_ = SaveJob(job)

	job.MarkDone()
	assert.NoError(t, SaveJob(job))

	jobs, _ := GetJobs()
	assert.Len(t, jobs, 1)
	assert.True(t, jobs[0].IsDone())
}

func TestDispatcher_GetJobs_EmptyStore(t *testing.T) {
	setupTempStorage(t)
	jobs, err := GetJobs()
	assert.NoError(t, err)
	assert.Empty(t, jobs)
}

func TestDispatcher_RemoveJob(t *testing.T) {
	setupTempStorage(t)

	keep, _ := model.CreateJob("keep")
	drop, _ := model.CreateJob("drop")
	_ = SaveJob(keep)
	_ = SaveJob(drop)

	assert.NoError(t, RemoveJob(drop.GetID()))

	jobs, _ := GetJobs()
	assert.Len(t, jobs, 1)
	assert.Equal(t, keep.GetID(), jobs[0].GetID())
}

func TestDispatcher_RemoveJob_NotFound(t *testing.T) {
	setupTempStorage(t)
	err := RemoveJob(uuid.New())
	assert.Equal(t, util.ErrNotFound, err)
}

func TestDispatcher_GetJobs_PreservesFields(t *testing.T) {
	setupTempStorage(t)

	job, _ := model.CreateJob("field-check")
	job.MarkDone()
	_ = SaveJob(job)

	jobs, _ := GetJobs()
	assert.Equal(t, "field-check", jobs[0].GetName())
	assert.True(t, jobs[0].IsDone())
	assert.Equal(t, job.GetID(), jobs[0].GetID())
}

// =============================================================================
// Task — SaveTask / GetTasks / RemoveTask
// =============================================================================

func TestDispatcher_SaveAndGetTask(t *testing.T) {
	setupTempStorage(t)

	task, _ := model.CreateTask("write tests", []string{"home", "projects"})
	assert.NoError(t, SaveTask(task))

	tasks, err := GetTasks()
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, task.GetID(), tasks[0].GetID())
	assert.Equal(t, "write tests", tasks[0].GetName())
	assert.Equal(t, []string{"home", "projects"}, tasks[0].GetPath())
}

func TestDispatcher_SaveTask_UpdatesExisting(t *testing.T) {
	setupTempStorage(t)

	task, _ := model.CreateTask("pending", []string{"root"})
	_ = SaveTask(task)

	task.MarkDone()
	assert.NoError(t, SaveTask(task))

	tasks, _ := GetTasks()
	assert.Len(t, tasks, 1)
	assert.True(t, tasks[0].IsDone())
}

func TestDispatcher_SaveTask_PreservesEpicID(t *testing.T) {
	setupTempStorage(t)

	epicID := uuid.New()
	task, _ := model.CreateTask("linked", []string{"root"}, epicID)
	_ = SaveTask(task)

	tasks, _ := GetTasks()
	assert.NotNil(t, tasks[0].GetEpicID())
	assert.Equal(t, epicID, *tasks[0].GetEpicID())
}

func TestDispatcher_GetTasks_EmptyStore(t *testing.T) {
	setupTempStorage(t)
	tasks, err := GetTasks()
	assert.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestDispatcher_RemoveTask(t *testing.T) {
	setupTempStorage(t)

	keep, _ := model.CreateTask("keep", []string{"root"})
	drop, _ := model.CreateTask("drop", []string{"root"})
	_ = SaveTask(keep)
	_ = SaveTask(drop)

	assert.NoError(t, RemoveTask(drop.GetID()))

	tasks, _ := GetTasks()
	assert.Len(t, tasks, 1)
	assert.Equal(t, keep.GetID(), tasks[0].GetID())
}

func TestDispatcher_RemoveTask_NotFound(t *testing.T) {
	setupTempStorage(t)
	err := RemoveTask(uuid.New())
	assert.Equal(t, util.ErrNotFound, err)
}

// =============================================================================
// Epic — SaveEpic / GetEpics / RemoveEpic
// =============================================================================

func TestDispatcher_SaveAndGetEpic(t *testing.T) {
	setupTempStorage(t)

	taskID := uuid.New()
	epic, _ := model.CreateEpic("v2 release", []string{"home", "projects"}, []uuid.UUID{taskID})
	assert.NoError(t, SaveEpic(epic))

	epics, err := GetEpics()
	assert.NoError(t, err)
	assert.Len(t, epics, 1)
	assert.Equal(t, epic.GetID(), epics[0].GetID())
	assert.Equal(t, "v2 release", epics[0].GetName())
	assert.Len(t, epics[0].GetTasksID(), 1)
	assert.Equal(t, taskID, epics[0].GetTasksID()[0])
}

func TestDispatcher_SaveEpic_UpdatesExisting(t *testing.T) {
	setupTempStorage(t)

	epic, _ := model.CreateEpic("epic", []string{"root"}, nil)
	_ = SaveEpic(epic)

	epic.MarkDone()
	assert.NoError(t, SaveEpic(epic))

	epics, _ := GetEpics()
	assert.Len(t, epics, 1)
	assert.True(t, epics[0].IsDone())
}

func TestDispatcher_GetEpics_EmptyStore(t *testing.T) {
	setupTempStorage(t)
	epics, err := GetEpics()
	assert.NoError(t, err)
	assert.Empty(t, epics)
}

func TestDispatcher_RemoveEpic(t *testing.T) {
	setupTempStorage(t)

	keep, _ := model.CreateEpic("keep", []string{"root"}, nil)
	drop, _ := model.CreateEpic("drop", []string{"root"}, nil)
	_ = SaveEpic(keep)
	_ = SaveEpic(drop)

	assert.NoError(t, RemoveEpic(drop.GetID()))

	epics, _ := GetEpics()
	assert.Len(t, epics, 1)
	assert.Equal(t, keep.GetID(), epics[0].GetID())
}

func TestDispatcher_RemoveEpic_NotFound(t *testing.T) {
	setupTempStorage(t)
	err := RemoveEpic(uuid.New())
	assert.Equal(t, util.ErrNotFound, err)
}

// =============================================================================
// Cross-entity isolation
// =============================================================================

func TestDispatcher_EntitiesAreIsolated(t *testing.T) {
	setupTempStorage(t)

	job, _ := model.CreateJob("a job")
	task, _ := model.CreateTask("a task", []string{"root"})
	epic, _ := model.CreateEpic("an epic", []string{"root"}, nil)

	_ = SaveJob(job)
	_ = SaveTask(task)
	_ = SaveEpic(epic)

	jobs, _ := GetJobs()
	tasks, _ := GetTasks()
	epics, _ := GetEpics()

	assert.Len(t, jobs, 1)
	assert.Len(t, tasks, 1)
	assert.Len(t, epics, 1)
}

func TestDispatcher_MultipleEntities(t *testing.T) {
	setupTempStorage(t)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		job, _ := model.CreateJob(name)
		_ = SaveJob(job)
	}
	jobs, _ := GetJobs()
	assert.Len(t, jobs, 3)
}
