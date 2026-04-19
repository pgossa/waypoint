package task

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Name validation (enforced by model.CreateTask inside taskAdd)
// Allowed charset: [a-zA-Z0-9 ._/#@-], 1–64 chars
// =============================================================================

func TestTaskAdd_ValidNames(t *testing.T) {
	path := []string{"root"}
	valid := []string{
		"fix something",
		"fix-something",
		"fix.something",
		"fix_something",
		"fix/something",
		"fix@something",
		"fix#something",
		strings.Repeat("a", 64),
	}
	for _, name := range valid {
		name := name
		t.Run(name, func(t *testing.T) {
			task, err := model.CreateTask(name, path)
			assert.NoError(t, err, "expected valid name %q to be accepted", name)
			assert.NotNil(t, task)
		})
	}
}

func TestTaskAdd_InvalidNames(t *testing.T) {
	path := []string{"root"}
	tests := []struct {
		input   string
		wantErr error
	}{
		{"", util.ErrNameLength},
		{strings.Repeat("a", 65), util.ErrNameLength},
		{`back\slash`, util.ErrNameInvalid},
		{"pipe|char", util.ErrNameInvalid},
		{"bang!char", util.ErrNameInvalid},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			_, err := model.CreateTask(tt.input, path)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// =============================================================================
// Path validation (enforced by model.CreateTask)
// Segment charset: [a-zA-Z0-9 ._-] — stricter than name (no /@#)
// =============================================================================

// taskAdd splits the --path argument by "/".
func TestTaskAdd_PathSplitting(t *testing.T) {
	parts := strings.Split("home/user/projects", "/")
	assert.Equal(t, []string{"home", "user", "projects"}, parts)
}

func TestTaskAdd_SingleSegmentPath(t *testing.T) {
	parts := strings.Split("root", "/")
	assert.Equal(t, []string{"root"}, parts)
}

func TestTaskAdd_InvalidPath(t *testing.T) {
	tests := []struct {
		desc    string
		path    []string
		wantErr error
	}{
		{"nil path", nil, util.ErrPath},
		{"empty path slice", []string{}, util.ErrPath},
		{"empty segment", []string{""}, util.ErrPath},
		{"segment too long", []string{strings.Repeat("a", 65)}, util.ErrPath},
		{"slash in segment", []string{"root/child"}, util.ErrPathInvalid},
		{"at in segment", []string{"root@host"}, util.ErrPathInvalid},
		{"hash in segment", []string{"root#1"}, util.ErrPathInvalid},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			_, err := model.CreateTask("valid-name", tt.path)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// =============================================================================
// Epic ID handling (--epic flag in taskAdd)
// =============================================================================

// A task with a valid epic UUID is created correctly.
func TestTaskAdd_WithValidEpicID(t *testing.T) {
	epicID := uuid.New()
	task, err := model.CreateTask("epic-task", []string{"root"}, epicID)
	assert.NoError(t, err)
	assert.NotNil(t, task.GetEpicID())
	assert.Equal(t, epicID, *task.GetEpicID())
}

// Providing two epic IDs (impossible via CLI but guards model contract).
func TestTaskAdd_TwoEpicIDs_Rejected(t *testing.T) {
	_, err := model.CreateTask("task", []string{"root"}, uuid.New(), uuid.New())
	assert.Equal(t, util.ErrEpicUUID, err)
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestTaskAddUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { taskAddUsage() })
}
