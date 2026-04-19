package storage

import (
	"testing"

	"waypoint/config"
	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
)

// setupTempStorage redirects HOME to a temp dir so all storage I/O is isolated.
// It also resets the config cache so Load() picks up the new HOME.
func setupTempStorage(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	config.Reset()
	t.Cleanup(config.Reset)
}

// =============================================================================
// Job CRUD
// =============================================================================

func TestJobJSON_SaveAndGet(t *testing.T) {
	setupTempStorage(t)

	job, _ := model.CreateJob("buy milk")
	if err := upsertJobJSON(job); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	jobs, err := getJobsJSON()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Name != "buy milk" {
		t.Errorf("name mismatch: got %q", jobs[0].Name)
	}
	if jobs[0].ID != job.GetID().String() {
		t.Errorf("ID mismatch")
	}
}

func TestJobJSON_GetFromEmptyStore(t *testing.T) {
	setupTempStorage(t)

	jobs, err := getJobsJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected empty store, got %d jobs", len(jobs))
	}
}

func TestJobJSON_Upsert_UpdatesExisting(t *testing.T) {
	setupTempStorage(t)

	job, _ := model.CreateJob("original-name")
	_ = upsertJobJSON(job)

	// mutate and upsert again
	job.MarkDone()
	if err := upsertJobJSON(job); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	jobs, _ := getJobsJSON()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job after upsert, got %d", len(jobs))
	}
	if !jobs[0].Done {
		t.Error("expected done=true after upsert update")
	}
}

func TestJobJSON_SaveMultiple(t *testing.T) {
	setupTempStorage(t)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		job, _ := model.CreateJob(name)
		_ = upsertJobJSON(job)
	}

	jobs, _ := getJobsJSON()
	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs, got %d", len(jobs))
	}
}

func TestJobJSON_Remove(t *testing.T) {
	setupTempStorage(t)

	a, _ := model.CreateJob("keep-me")
	b, _ := model.CreateJob("remove-me")
	_ = upsertJobJSON(a)
	_ = upsertJobJSON(b)

	if err := removeJobJSON(b.GetID()); err != nil {
		t.Fatalf("remove: %v", err)
	}

	jobs, _ := getJobsJSON()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job after remove, got %d", len(jobs))
	}
	if jobs[0].Name != "keep-me" {
		t.Errorf("wrong job remained: %q", jobs[0].Name)
	}
}

func TestJobJSON_Remove_NotFound(t *testing.T) {
	setupTempStorage(t)

	err := removeJobJSON(uuid.New())
	if err != util.ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

// =============================================================================
// Task CRUD
// =============================================================================

func TestTaskJSON_SaveAndGet(t *testing.T) {
	setupTempStorage(t)

	task, _ := model.CreateTask("write tests", []string{"home", "projects"})
	if err := upsertTaskJSON(task); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	tasks, err := getTasksJSON()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "write tests" {
		t.Errorf("name mismatch: got %q", tasks[0].Name)
	}
	if len(tasks[0].Path) != 2 || tasks[0].Path[0] != "home" {
		t.Errorf("path mismatch: got %v", tasks[0].Path)
	}
}

func TestTaskJSON_Upsert_UpdatesDoneFlag(t *testing.T) {
	setupTempStorage(t)

	task, _ := model.CreateTask("pending", []string{"root"})
	_ = upsertTaskJSON(task)

	task.MarkDone()
	_ = upsertTaskJSON(task)

	tasks, _ := getTasksJSON()
	if len(tasks) != 1 || !tasks[0].Done {
		t.Error("expected done=true after upsert update")
	}
}

func TestTaskJSON_WithEpicID_Preserved(t *testing.T) {
	setupTempStorage(t)

	epicID := uuid.New()
	task, _ := model.CreateTask("linked", []string{"root"}, epicID)
	_ = upsertTaskJSON(task)

	tasks, _ := getTasksJSON()
	if tasks[0].EpicID == nil || *tasks[0].EpicID != epicID.String() {
		t.Errorf("epicID not preserved, got %v", tasks[0].EpicID)
	}
}

func TestTaskJSON_Remove(t *testing.T) {
	setupTempStorage(t)

	keep, _ := model.CreateTask("keep", []string{"root"})
	drop, _ := model.CreateTask("drop", []string{"root"})
	_ = upsertTaskJSON(keep)
	_ = upsertTaskJSON(drop)

	if err := removeTaskJSON(drop.GetID()); err != nil {
		t.Fatalf("remove: %v", err)
	}

	tasks, _ := getTasksJSON()
	if len(tasks) != 1 || tasks[0].Name != "keep" {
		t.Errorf("unexpected tasks after remove: %v", tasks)
	}
}

func TestTaskJSON_Remove_NotFound(t *testing.T) {
	setupTempStorage(t)
	err := removeTaskJSON(uuid.New())
	if err != util.ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

// =============================================================================
// Epic CRUD
// =============================================================================

func TestEpicJSON_SaveAndGet(t *testing.T) {
	setupTempStorage(t)

	taskID := uuid.New()
	epic, _ := model.CreateEpic("v2 release", []string{"home", "projects"}, []uuid.UUID{taskID})
	if err := upsertEpicJSON(epic); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	epics, err := getEpicsJSON()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(epics) != 1 {
		t.Fatalf("expected 1 epic, got %d", len(epics))
	}
	if epics[0].Name != "v2 release" {
		t.Errorf("name mismatch: got %q", epics[0].Name)
	}
	if len(epics[0].TasksID) != 1 || epics[0].TasksID[0] != taskID.String() {
		t.Errorf("tasks_id mismatch: got %v", epics[0].TasksID)
	}
}

func TestEpicJSON_Upsert_UpdatesDoneFlag(t *testing.T) {
	setupTempStorage(t)

	epic, _ := model.CreateEpic("epic", []string{"root"}, nil)
	_ = upsertEpicJSON(epic)
	epic.MarkDone()
	_ = upsertEpicJSON(epic)

	epics, _ := getEpicsJSON()
	if len(epics) != 1 || !epics[0].Done {
		t.Error("expected done=true after upsert update")
	}
}

func TestEpicJSON_Remove(t *testing.T) {
	setupTempStorage(t)

	keep, _ := model.CreateEpic("keep", []string{"root"}, nil)
	drop, _ := model.CreateEpic("drop", []string{"root"}, nil)
	_ = upsertEpicJSON(keep)
	_ = upsertEpicJSON(drop)

	if err := removeEpicJSON(drop.GetID()); err != nil {
		t.Fatalf("remove: %v", err)
	}

	epics, _ := getEpicsJSON()
	if len(epics) != 1 || epics[0].Name != "keep" {
		t.Errorf("unexpected epics after remove: %v", epics)
	}
}

func TestEpicJSON_Remove_NotFound(t *testing.T) {
	setupTempStorage(t)
	err := removeEpicJSON(uuid.New())
	if err != util.ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

// =============================================================================
// Isolation — each entity type is stored independently
// =============================================================================

func TestJSON_JobsAndTasksAreIndependent(t *testing.T) {
	setupTempStorage(t)

	job, _ := model.CreateJob("a job")
	task, _ := model.CreateTask("a task", []string{"root"})
	_ = upsertJobJSON(job)
	_ = upsertTaskJSON(task)

	jobs, _ := getJobsJSON()
	tasks, _ := getTasksJSON()

	if len(jobs) != 1 || len(tasks) != 1 {
		t.Errorf("isolation broken: jobs=%d tasks=%d", len(jobs), len(tasks))
	}
}
