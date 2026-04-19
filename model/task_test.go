package model_test

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
)

// =============================================================================
// CreateTask
// =============================================================================

func TestCreateTask(t *testing.T) {
	validPath := []string{"root"}
	epicID := uuid.New()

	tests := []struct {
		name        string
		inputName   string
		inputPath   []string
		inputEpicID []uuid.UUID
		wantErr     error
	}{
		{"standalone task", "fix something", validPath, nil, nil},
		{"task with epic", "fix something", validPath, []uuid.UUID{epicID}, nil},
		{"too many epics", "fix something", validPath, []uuid.UUID{epicID, uuid.New()}, util.ErrEpicUUID},
		{"empty name", "", validPath, nil, util.ErrNameLength},
		{"name too long", strings.Repeat("a", 65), validPath, nil, util.ErrNameLength},
		{"name at max length", strings.Repeat("a", 64), validPath, nil, nil},
		{"name with backslash", `fix\something`, validPath, nil, util.ErrNameInvalid},
		{"name with pipe", "fix|something", validPath, nil, util.ErrNameInvalid},
		{"nil path", "fix something", nil, nil, util.ErrPath},
		{"empty path slice", "fix something", []string{}, nil, util.ErrPath},
		{"empty segment", "fix something", []string{""}, nil, util.ErrPath},
		{"segment too long", "fix something", []string{strings.Repeat("a", 65)}, nil, util.ErrPath},
		{"segment with slash", "fix something", []string{"root/child"}, nil, util.ErrPathInvalid},
		{"multi-segment valid", "fix something", []string{"root", "child"}, nil, nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			task, err := model.CreateTask(tt.inputName, tt.inputPath, tt.inputEpicID...)
			if err != tt.wantErr {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if err == nil && task == nil {
				t.Error("expected task, got nil")
			}
		})
	}
}

func TestCreateTask_NewIDEachCall(t *testing.T) {
	a, _ := model.CreateTask("task-a", []string{"root"})
	b, _ := model.CreateTask("task-b", []string{"root"})
	if a.GetID() == b.GetID() {
		t.Error("two CreateTask calls should produce distinct IDs")
	}
}

func TestCreateTask_StartsNotDone(t *testing.T) {
	task, _ := model.CreateTask("task", []string{"root"})
	if task.IsDone() {
		t.Error("new task should not be done")
	}
}

// =============================================================================
// MarkDone / IsDone
// =============================================================================

func TestTask_MarkDone(t *testing.T) {
	task, _ := model.CreateTask("fix something", []string{"root"})
	if task.IsDone() {
		t.Error("expected not done on creation")
	}
	task.MarkDone()
	if !task.IsDone() {
		t.Error("expected done after MarkDone()")
	}
}

// =============================================================================
// EpicID management
// =============================================================================

func TestTask_NoEpicByDefault(t *testing.T) {
	task, _ := model.CreateTask("standalone", []string{"root"})
	if task.GetEpicID() != nil {
		t.Error("new task without epic should have nil epicID")
	}
}

func TestTask_SetEpicID(t *testing.T) {
	task, _ := model.CreateTask("task", []string{"root"})
	epicID := uuid.New()
	task.SetEpicID(epicID)

	got := task.GetEpicID()
	if got == nil {
		t.Fatal("expected epicID to be set, got nil")
	}
	if *got != epicID {
		t.Errorf("got epicID %v, want %v", *got, epicID)
	}
}

func TestTask_UnsetEpicID(t *testing.T) {
	epicID := uuid.New()
	task, _ := model.CreateTask("task", []string{"root"}, epicID)
	task.UnsetEpicID()
	if task.GetEpicID() != nil {
		t.Error("expected nil epicID after UnsetEpicID()")
	}
}

func TestTask_SetEpicID_Overwrites(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	task, _ := model.CreateTask("task", []string{"root"}, first)
	task.SetEpicID(second)
	got := task.GetEpicID()
	if got == nil || *got != second {
		t.Errorf("expected epicID to be updated to %v", second)
	}
}

// =============================================================================
// ToJSON
// =============================================================================

func TestTask_ToJSON_StandaloneFields(t *testing.T) {
	path := []string{"home", "user"}
	task, _ := model.CreateTask("fix something", path)
	j := task.ToJSON()

	if j.Name != "fix something" {
		t.Errorf("name mismatch: got %q", j.Name)
	}
	if j.Done {
		t.Error("expected done=false on new task")
	}
	if j.EpicID != nil {
		t.Error("expected nil epic_id for standalone task")
	}
	if len(j.Path) != 2 || j.Path[0] != "home" || j.Path[1] != "user" {
		t.Errorf("path mismatch: got %v", j.Path)
	}
	if _, err := uuid.Parse(j.ID); err != nil {
		t.Errorf("expected valid UUID, got %q", j.ID)
	}
}

func TestTask_ToJSON_WithEpicID(t *testing.T) {
	epicID := uuid.New()
	task, _ := model.CreateTask("fix something", []string{"root"}, epicID)
	j := task.ToJSON()

	if j.EpicID == nil {
		t.Fatal("expected epic_id to be set")
	}
	if *j.EpicID != epicID.String() {
		t.Errorf("epic_id mismatch: got %q, want %q", *j.EpicID, epicID.String())
	}
}

// =============================================================================
// FromTaskJSON — round-trip
// =============================================================================

func TestTask_RoundTrip_Standalone(t *testing.T) {
	original, _ := model.CreateTask("round-trip", []string{"home", "user"})
	original.MarkDone()

	restored, err := model.FromTaskJSON(original.ToJSON())
	if err != nil {
		t.Fatalf("FromTaskJSON error: %v", err)
	}

	if restored.GetID() != original.GetID() {
		t.Errorf("ID mismatch")
	}
	if restored.GetName() != original.GetName() {
		t.Errorf("name mismatch")
	}
	if restored.IsDone() != original.IsDone() {
		t.Errorf("done mismatch")
	}
	if len(restored.GetPath()) != len(original.GetPath()) {
		t.Errorf("path length mismatch")
	}
	if restored.GetEpicID() != nil {
		t.Error("expected nil epicID after round-trip")
	}
}

func TestTask_RoundTrip_WithEpic(t *testing.T) {
	epicID := uuid.New()
	original, _ := model.CreateTask("epic-task", []string{"root"}, epicID)

	restored, err := model.FromTaskJSON(original.ToJSON())
	if err != nil {
		t.Fatalf("FromTaskJSON error: %v", err)
	}

	got := restored.GetEpicID()
	if got == nil {
		t.Fatal("expected epicID preserved through round-trip")
	}
	if *got != epicID {
		t.Errorf("epicID mismatch: got %v, want %v", *got, epicID)
	}
}

func TestFromTaskJSON_InvalidTaskUUID(t *testing.T) {
	j := model.TaskJSON{ID: "bad", Name: "x", Path: []string{"root"}}
	_, err := model.FromTaskJSON(j)
	if err == nil {
		t.Error("expected error for invalid task UUID")
	}
}

func TestFromTaskJSON_InvalidEpicUUID(t *testing.T) {
	bad := "not-a-uuid"
	j := model.TaskJSON{ID: uuid.New().String(), Name: "x", Path: []string{"root"}, EpicID: &bad}
	_, err := model.FromTaskJSON(j)
	if err == nil {
		t.Error("expected error for invalid epic UUID")
	}
}
