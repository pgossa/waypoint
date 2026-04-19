package model_test

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
)

// =============================================================================
// CreateEpic
// =============================================================================

func TestCreateEpic(t *testing.T) {
	validPath := []string{"root"}
	validTasks := []uuid.UUID{uuid.New()}

	tests := []struct {
		name    string
		input   string
		path    []string
		tasksID []uuid.UUID
		wantErr error
	}{
		{"valid, no tasks", "fix something", validPath, nil, nil},
		{"valid, with tasks", "fix something", validPath, validTasks, nil},
		{"empty name", "", validPath, nil, util.ErrNameLength},
		{"name too long", strings.Repeat("a", 65), validPath, nil, util.ErrNameLength},
		{"name at max length", strings.Repeat("a", 64), validPath, nil, nil},
		{"name invalid chars", "fix|something", validPath, nil, util.ErrNameInvalid},
		{"nil path", "fix something", nil, nil, util.ErrPath},
		{"empty path slice", "fix something", []string{}, nil, util.ErrPath},
		{"empty path segment", "fix something", []string{""}, nil, util.ErrPath},
		{"segment too long", "fix something", []string{strings.Repeat("a", 65)}, nil, util.ErrPath},
		{"segment invalid chars", "fix something", []string{"root|child"}, nil, util.ErrPathInvalid},
		{"multi segment valid", "fix something", []string{"root", "child"}, nil, nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			epic, err := model.CreateEpic(tt.input, tt.path, tt.tasksID)
			if err != tt.wantErr {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if err == nil && epic == nil {
				t.Error("expected epic, got nil")
			}
		})
	}
}

func TestCreateEpic_NewIDEachCall(t *testing.T) {
	a, _ := model.CreateEpic("epic-a", []string{"root"}, nil)
	b, _ := model.CreateEpic("epic-b", []string{"root"}, nil)
	if a.GetID() == b.GetID() {
		t.Error("two CreateEpic calls should produce distinct IDs")
	}
}

func TestCreateEpic_StartsNotDone(t *testing.T) {
	epic, _ := model.CreateEpic("epic", []string{"root"}, nil)
	if epic.IsDone() {
		t.Error("new epic should not be done")
	}
}

// =============================================================================
// MarkDone / IsDone
// =============================================================================

func TestEpic_MarkDone(t *testing.T) {
	epic, _ := model.CreateEpic("fix something", []string{"root"}, nil)
	if epic.IsDone() {
		t.Error("expected not done on creation")
	}
	epic.MarkDone()
	if !epic.IsDone() {
		t.Error("expected done after MarkDone()")
	}
}

// =============================================================================
// AddTaskID / RemoveTaskID / GetTasksID
// =============================================================================

func TestEpic_AddTaskID(t *testing.T) {
	epic, _ := model.CreateEpic("epic", []string{"root"}, nil)
	taskID := uuid.New()

	if err := epic.AddTaskID(taskID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := epic.GetTasksID()
	if len(ids) != 1 || ids[0] != taskID {
		t.Errorf("expected [%v], got %v", taskID, ids)
	}
}

func TestEpic_AddTaskID_Duplicate(t *testing.T) {
	epic, _ := model.CreateEpic("epic", []string{"root"}, nil)
	taskID := uuid.New()

	_ = epic.AddTaskID(taskID)
	err := epic.AddTaskID(taskID)
	if err != util.ErrDuplicateTask {
		t.Errorf("got error %v, want ErrDuplicateTask", err)
	}
}

func TestEpic_AddMultipleTaskIDs(t *testing.T) {
	epic, _ := model.CreateEpic("epic", []string{"root"}, nil)
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	for _, id := range ids {
		if err := epic.AddTaskID(id); err != nil {
			t.Fatalf("unexpected error adding %v: %v", id, err)
		}
	}
	if len(epic.GetTasksID()) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(epic.GetTasksID()))
	}
}

func TestEpic_RemoveTaskID_Existing(t *testing.T) {
	taskA, taskB := uuid.New(), uuid.New()
	epic, _ := model.CreateEpic("epic", []string{"root"}, []uuid.UUID{taskA, taskB})

	epic.RemoveTaskID(taskA)

	ids := epic.GetTasksID()
	if len(ids) != 1 || ids[0] != taskB {
		t.Errorf("expected only taskB remaining, got %v", ids)
	}
}

func TestEpic_RemoveTaskID_NonExistent_NoEffect(t *testing.T) {
	taskA := uuid.New()
	epic, _ := model.CreateEpic("epic", []string{"root"}, []uuid.UUID{taskA})

	epic.RemoveTaskID(uuid.New()) // different UUID

	ids := epic.GetTasksID()
	if len(ids) != 1 || ids[0] != taskA {
		t.Errorf("expected taskA to remain, got %v", ids)
	}
}

func TestEpic_RemoveTaskID_Last(t *testing.T) {
	taskID := uuid.New()
	epic, _ := model.CreateEpic("epic", []string{"root"}, []uuid.UUID{taskID})
	epic.RemoveTaskID(taskID)
	if len(epic.GetTasksID()) != 0 {
		t.Error("expected empty tasks after removing the only task")
	}
}

// =============================================================================
// ToJSON
// =============================================================================

func TestEpic_ToJSON_NoTasks(t *testing.T) {
	epic, _ := model.CreateEpic("fix something", []string{"root"}, nil)
	j := epic.ToJSON()

	if j.Name != "fix something" {
		t.Errorf("name mismatch: got %q", j.Name)
	}
	if j.Done {
		t.Error("expected done=false")
	}
	if len(j.TasksID) != 0 {
		t.Errorf("expected empty tasks_id, got %v", j.TasksID)
	}
	if _, err := uuid.Parse(j.ID); err != nil {
		t.Errorf("invalid UUID: %q", j.ID)
	}
}

func TestEpic_ToJSON_WithTasks(t *testing.T) {
	taskID := uuid.New()
	epic, _ := model.CreateEpic("fix something", []string{"root"}, []uuid.UUID{taskID})
	j := epic.ToJSON()

	if len(j.TasksID) != 1 || j.TasksID[0] != taskID.String() {
		t.Errorf("tasks_id mismatch: got %v", j.TasksID)
	}
}

// =============================================================================
// FromEpicJSON — round-trip
// =============================================================================

func TestEpic_RoundTrip(t *testing.T) {
	taskA, taskB := uuid.New(), uuid.New()
	original, _ := model.CreateEpic("round-trip", []string{"home", "user"}, []uuid.UUID{taskA, taskB})
	original.MarkDone()

	restored, err := model.FromEpicJSON(original.ToJSON())
	if err != nil {
		t.Fatalf("FromEpicJSON error: %v", err)
	}

	if restored.GetID() != original.GetID() {
		t.Error("ID mismatch")
	}
	if restored.GetName() != original.GetName() {
		t.Error("name mismatch")
	}
	if restored.IsDone() != original.IsDone() {
		t.Error("done mismatch")
	}

	restoredIDs := restored.GetTasksID()
	if len(restoredIDs) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(restoredIDs))
	}
	// order is preserved
	if restoredIDs[0] != taskA || restoredIDs[1] != taskB {
		t.Errorf("task IDs mismatch: got %v", restoredIDs)
	}
}

func TestFromEpicJSON_InvalidEpicUUID(t *testing.T) {
	j := model.EpicJSON{ID: "bad-uuid", Name: "x", Path: []string{"root"}}
	_, err := model.FromEpicJSON(j)
	if err == nil {
		t.Error("expected error for invalid epic UUID")
	}
}

func TestFromEpicJSON_InvalidTaskUUID(t *testing.T) {
	j := model.EpicJSON{
		ID:      uuid.New().String(),
		Name:    "x",
		Path:    []string{"root"},
		TasksID: []string{"not-a-uuid"},
	}
	_, err := model.FromEpicJSON(j)
	if err == nil {
		t.Error("expected error for invalid task UUID in epic")
	}
}
