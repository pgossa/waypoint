package epic

import (
	"testing"

	"waypoint/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// epicRemove locates the epic via FindEpic, shows a huh confirmation
// (untestable in unit tests), optionally removes linked subtasks, then
// removes the epic. We test the FindEpic resolution and linked-task detection.

// =============================================================================
// FindEpic resolution (drives epicRemove branching)
// =============================================================================

func TestEpicRemove_NoMatchFound(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "v2-release"),
		newEpic(t, "v3-release"),
	}
	idx, matches := FindEpic("nonexistent", epics)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

func TestEpicRemove_SingleMatchByName(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "v2-release"),
		newEpic(t, "v3-release"),
	}
	idx, matches := FindEpic("v2-release", epics)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestEpicRemove_SingleMatchByPartialName(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "alpha-release"),
		newEpic(t, "beta-release"),
	}
	idx, matches := FindEpic("alpha", epics)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestEpicRemove_MultipleMatches(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "v2-backend"),
		newEpic(t, "v2-frontend"),
		newEpic(t, "v3-backend"),
	}
	idx, matches := FindEpic("v2", epics)
	assert.Equal(t, -1, idx)
	assert.ElementsMatch(t, []int{0, 1}, matches)
}

func TestEpicRemove_ByIndex(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "alpha"),
		newEpic(t, "beta"),
		newEpic(t, "gamma"),
	}
	idx, matches := FindEpic("2", epics)
	assert.Equal(t, 1, idx)
	assert.Equal(t, []int{1}, matches)
}

// =============================================================================
// Linked subtask detection
// epicRemove asks for confirmation before removing subtasks when len > 0.
// =============================================================================

func TestEpicRemove_EpicWithSubtasks_Detectable(t *testing.T) {
	task, _ := model.CreateTask("subtask", []string{"root"})
	epic, _ := model.CreateEpic("release", []string{"root"}, []uuid.UUID{task.GetID()})
	assert.Greater(t, len(epic.GetTasksID()), 0)
}

func TestEpicRemove_EpicWithNoSubtasks_NoConfirmNeeded(t *testing.T) {
	epic := newEpic(t, "empty-epic")
	assert.Empty(t, epic.GetTasksID())
}

// =============================================================================
// Epic identity
// =============================================================================

func TestEpicRemove_EpicHasStableID(t *testing.T) {
	epic := newEpic(t, "removable")
	assert.Equal(t, epic.GetID(), epic.GetID())
	assert.NotEmpty(t, epic.GetID())
}

func TestEpicRemove_DifferentEpics_DifferentIDs(t *testing.T) {
	a := newEpic(t, "epic-a")
	b := newEpic(t, "epic-b")
	assert.NotEqual(t, a.GetID(), b.GetID())
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestEpicRemoveUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { epicRemoveUsage() })
}
