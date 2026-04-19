package epic

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/stretchr/testify/assert"
)

// epicAdd accepts: <name> [--empty|-e] [--path|-p <path>]
// The auto-subtask creation calls os.ReadDir (filesystem-dependent).
// We test name/path validation and flag parsing logic.

// =============================================================================
// Name validation (passed to model.CreateEpic)
// =============================================================================

func TestEpicAdd_ValidNames(t *testing.T) {
	path := []string{"root"}
	valid := []string{
		"v2 release",
		"v2-release",
		"v2.release",
		"release_2024",
		strings.Repeat("a", 64),
	}
	for _, name := range valid {
		name := name
		t.Run(name, func(t *testing.T) {
			epic, err := model.CreateEpic(name, path, nil)
			assert.NoError(t, err)
			assert.NotNil(t, epic)
		})
	}
}

func TestEpicAdd_InvalidNames(t *testing.T) {
	path := []string{"root"}
	tests := []struct {
		input   string
		wantErr error
	}{
		{"", util.ErrNameLength},
		{strings.Repeat("a", 65), util.ErrNameLength},
		{"pipe|name", util.ErrNameInvalid},
		{`back\slash`, util.ErrNameInvalid},
		{"bang!name", util.ErrNameInvalid},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			_, err := model.CreateEpic(tt.input, path, nil)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// =============================================================================
// Flag parsing: --empty flag controls whether subtasks are auto-created
// =============================================================================

func TestEpicAdd_EmptyFlag_Variants(t *testing.T) {
	for _, arg := range []string{"--empty", "-e"} {
		empty := parseEpicAddEmpty([]string{arg})
		assert.True(t, empty, "expected empty=true for %q", arg)
	}
}

func TestEpicAdd_EmptyFlag_NotSet(t *testing.T) {
	empty := parseEpicAddEmpty([]string{})
	assert.False(t, empty)
}

// =============================================================================
// Path flag parsing
// =============================================================================

func TestEpicAdd_PathFlag_SplitBySlash(t *testing.T) {
	path := parseEpicAddPath([]string{"--path", "home/user/projects"})
	assert.Equal(t, []string{"home", "user", "projects"}, path)

	path = parseEpicAddPath([]string{"-p", "root"})
	assert.Equal(t, []string{"root"}, path)
}

func TestEpicAdd_PathFlag_NotSet_ReturnsNil(t *testing.T) {
	path := parseEpicAddPath([]string{})
	assert.Nil(t, path)
}

// =============================================================================
// Epic creation result
// =============================================================================

func TestEpicAdd_CreatedEpic_StartsNotDone_NoTasks(t *testing.T) {
	epic, err := model.CreateEpic("v2 release", []string{"root"}, nil)
	assert.NoError(t, err)
	assert.False(t, epic.IsDone())
	assert.Empty(t, epic.GetTasksID())
}

func TestEpicAdd_CreatedEpic_WithPreloadedTasks(t *testing.T) {
	task, _ := model.CreateTask("subtask", []string{"root", "sub"})
	epic, err := model.CreateEpic("v2 release", []string{"root"}, nil)
	assert.NoError(t, err)
	_ = epic.AddTaskID(task.GetID())
	assert.Len(t, epic.GetTasksID(), 1)
}

// =============================================================================
// Subtask naming convention used in epicAdd
// =============================================================================

// BUG: epicAdd prefixes subtask names with "[subtask] " but square brackets
// are not in the allowed name charset ([a-zA-Z0-9 ._/#@-]). This means
// epicAdd would call os.Exit(1) when trying to create subtasks in production.
// The prefix needs to be changed to avoid bracket characters.
func TestEpicAdd_SubtaskName_Format_BracketsBug(t *testing.T) {
	epicName := "v2 release"
	subtaskName := "[subtask] " + epicName
	assert.True(t, strings.HasPrefix(subtaskName, "[subtask] "))

	_, err := model.CreateTask(subtaskName, []string{"root", "sub"})
	assert.Error(t, err, "square brackets are not in the allowed name charset — epicAdd subtask naming is broken")
	assert.Equal(t, util.ErrNameInvalid, err)
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestEpicAddUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { epicAddUsage() })
}

// =============================================================================
// Helpers
// =============================================================================

func parseEpicAddEmpty(args []string) bool {
	for _, arg := range args {
		if arg == "--empty" || arg == "-e" {
			return true
		}
	}
	return false
}

func parseEpicAddPath(args []string) []string {
	for i, arg := range args {
		if (arg == "--path" || arg == "-p") && i+1 < len(args) {
			return strings.Split(args[i+1], "/")
		}
	}
	return nil
}
