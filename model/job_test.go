package model_test

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
)

// =============================================================================
// CreateJob
// =============================================================================

func TestCreateJob(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		wantErr   error
	}{
		{"valid name", "do something", nil},
		{"empty name", "", util.ErrNameLength},
		{"65 chars", strings.Repeat("a", 65), util.ErrNameLength},
		{"64 chars", strings.Repeat("a", 64), nil},
		{"with dash", "do-something", nil},
		{"with slash", "do/something", nil},
		{"with @", "do@something", nil},
		{"with #", "do#something", nil},
		{"backslash invalid", `do\this`, util.ErrNameInvalid},
		{"pipe invalid", "do|this", util.ErrNameInvalid},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			job, err := model.CreateJob(tt.inputName)
			if err != tt.wantErr {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if err == nil && job == nil {
				t.Error("expected job, got nil")
			}
		})
	}
}

func TestCreateJob_NewIDEachCall(t *testing.T) {
	a, _ := model.CreateJob("job-a")
	b, _ := model.CreateJob("job-b")
	if a.GetID() == b.GetID() {
		t.Error("two CreateJob calls should produce distinct IDs")
	}
}

func TestCreateJob_StartsNotDone(t *testing.T) {
	job, _ := model.CreateJob("something")
	if job.IsDone() {
		t.Error("new job should not be done")
	}
}

// =============================================================================
// MarkDone / IsDone
// =============================================================================

func TestJob_MarkDone(t *testing.T) {
	job, _ := model.CreateJob("do something")
	if job.IsDone() {
		t.Error("expected not done on creation")
	}
	job.MarkDone()
	if !job.IsDone() {
		t.Error("expected done after MarkDone()")
	}
}

func TestJob_MarkDone_Idempotent(t *testing.T) {
	job, _ := model.CreateJob("idempotent")
	job.MarkDone()
	job.MarkDone()
	if !job.IsDone() {
		t.Error("expected done after double MarkDone()")
	}
}

// =============================================================================
// ToJSON
// =============================================================================

func TestJob_ToJSON_Fields(t *testing.T) {
	job, _ := model.CreateJob("do something")
	j := job.ToJSON()

	if j.Name != "do something" {
		t.Errorf("got name %q, want %q", j.Name, "do something")
	}
	if j.Done {
		t.Error("expected done=false on new job")
	}
	if _, err := uuid.Parse(j.ID); err != nil {
		t.Errorf("expected valid UUID, got %q", j.ID)
	}
}

func TestJob_ToJSON_DoneFieldPreserved(t *testing.T) {
	job, _ := model.CreateJob("done-job")
	job.MarkDone()
	j := job.ToJSON()
	if !j.Done {
		t.Error("expected done=true in JSON")
	}
}

// =============================================================================
// FromJobJSON — round-trip
// =============================================================================

func TestJob_RoundTrip(t *testing.T) {
	original, _ := model.CreateJob("round-trip")
	original.MarkDone()

	j := original.ToJSON()
	restored, err := j.FromJobJSON()
	if err != nil {
		t.Fatalf("FromJobJSON error: %v", err)
	}

	if restored.GetID() != original.GetID() {
		t.Errorf("ID mismatch: got %v, want %v", restored.GetID(), original.GetID())
	}
	if restored.GetName() != original.GetName() {
		t.Errorf("name mismatch: got %q, want %q", restored.GetName(), original.GetName())
	}
	if restored.IsDone() != original.IsDone() {
		t.Errorf("done mismatch: got %v, want %v", restored.IsDone(), original.IsDone())
	}
}

func TestFromJobJSON_InvalidUUID(t *testing.T) {
	j := model.JobJSON{ID: "not-a-uuid", Name: "x", Done: false}
	_, err := j.FromJobJSON()
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}
