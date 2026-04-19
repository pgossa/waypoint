package epic

import (
	"testing"

	"waypoint/config"
	"waypoint/model"
	"waypoint/storage"

	"github.com/stretchr/testify/assert"
)

func setupEpicStorage(t *testing.T) {
	t.Helper()
	config.Reset()
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(config.Reset)
}

func newEpic(t *testing.T, name string) *model.Epic {
	t.Helper()
	epic, err := model.CreateEpic(name, []string{"root"}, nil)
	assert.NoError(t, err)
	return epic
}

// =============================================================================
// FindEpic — supports both numeric index (1-based) and name search
// =============================================================================

// In-range index returns (n-1, []int{n-1}).
func TestFindEpic_IndexInRange(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "alpha-release"),
		newEpic(t, "beta-release"),
		newEpic(t, "gamma-release"),
	}

	idx, matches := FindEpic("1", epics)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)

	idx, matches = FindEpic("3", epics)
	assert.Equal(t, 2, idx)
	assert.Equal(t, []int{2}, matches)
}

// Index 0 is below the 1-based minimum → (-1, nil).
func TestFindEpic_IndexZero_OutOfRange(t *testing.T) {
	epics := []*model.Epic{newEpic(t, "alpha")}
	idx, matches := FindEpic("0", epics)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// Index beyond list length → (-1, nil).
func TestFindEpic_IndexOutOfRange(t *testing.T) {
	epics := []*model.Epic{newEpic(t, "only")}
	idx, matches := FindEpic("99", epics)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// Negative index: Atoi succeeds but fails the >=1 guard → (-1, nil).
func TestFindEpic_NegativeIndex(t *testing.T) {
	epics := []*model.Epic{newEpic(t, "alpha")}
	idx, matches := FindEpic("-1", epics)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// Single exact name match.
func TestFindEpic_ExactNameMatch(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "alpha-release"),
		newEpic(t, "beta-release"),
	}
	idx, matches := FindEpic("alpha-release", epics)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

// Unambiguous partial name match.
func TestFindEpic_PartialName_SingleMatch(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "alpha-release"),
		newEpic(t, "beta-release"),
	}
	idx, matches := FindEpic("alpha", epics)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

// Ambiguous partial name match.
func TestFindEpic_PartialName_MultipleMatches(t *testing.T) {
	epics := []*model.Epic{
		newEpic(t, "v2-backend"),
		newEpic(t, "v2-frontend"),
		newEpic(t, "v3-backend"),
	}
	idx, matches := FindEpic("v2", epics)
	assert.Equal(t, -1, idx)
	assert.ElementsMatch(t, []int{0, 1}, matches)
}

// No match returns (-1, empty non-nil slice).
func TestFindEpic_NoMatch(t *testing.T) {
	epics := []*model.Epic{newEpic(t, "alpha-release")}
	idx, matches := FindEpic("nonexistent", epics)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
	assert.NotNil(t, matches)
}

func TestFindEpic_CaseInsensitive(t *testing.T) {
	epics := []*model.Epic{newEpic(t, "Alpha-Release")}

	idx, _ := FindEpic("alpha-release", epics)
	assert.Equal(t, 0, idx)

	idx, _ = FindEpic("ALPHA", epics)
	assert.Equal(t, 0, idx)
}

// Empty string matches every epic.
func TestFindEpic_EmptyInput_MatchesAll(t *testing.T) {
	epics := []*model.Epic{newEpic(t, "alpha"), newEpic(t, "beta")}
	idx, matches := FindEpic("", epics)
	assert.Equal(t, -1, idx)
	assert.Len(t, matches, 2)
}

func TestFindEpic_EmptyList(t *testing.T) {
	idx, matches := FindEpic("anything", []*model.Epic{})
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

// =============================================================================
// GetSortedEpics / GetActiveEpics — storage-integrated
// =============================================================================

func TestGetSortedEpics_ReturnsSorted(t *testing.T) {
	setupEpicStorage(t)

	for _, name := range []string{"gamma", "alpha", "beta"} {
		e, _ := model.CreateEpic(name, []string{"root"}, nil)
		assert.NoError(t, storage.SaveEpic(e))
	}

	epics, err := GetSortedEpics()
	assert.NoError(t, err)
	assert.Len(t, epics, 3)
	assert.Equal(t, "alpha", epics[0].GetName())
	assert.Equal(t, "beta", epics[1].GetName())
	assert.Equal(t, "gamma", epics[2].GetName())
}

func TestGetSortedEpics_Empty(t *testing.T) {
	setupEpicStorage(t)
	epics, err := GetSortedEpics()
	assert.NoError(t, err)
	assert.Empty(t, epics)
}

func TestGetActiveEpics_FiltersOutDone(t *testing.T) {
	setupEpicStorage(t)

	active, _ := model.CreateEpic("active", []string{"root"}, nil)
	done, _ := model.CreateEpic("done", []string{"root"}, nil)
	done.MarkDone()
	assert.NoError(t, storage.SaveEpic(active))
	assert.NoError(t, storage.SaveEpic(done))

	result, err := GetActiveEpics()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "active", result[0].GetName())
}

func TestGetActiveEpics_AllDone_ReturnsEmpty(t *testing.T) {
	setupEpicStorage(t)

	e, _ := model.CreateEpic("done", []string{"root"}, nil)
	e.MarkDone()
	assert.NoError(t, storage.SaveEpic(e))

	result, err := GetActiveEpics()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// =============================================================================
// PrintEpicWithIndex — smoke test (exercises formatting code paths)
// =============================================================================

func TestPrintEpicWithIndex_NoPanic(t *testing.T) {
	e := newEpic(t, "release")
	assert.NotPanics(t, func() { PrintEpicWithIndex(0, e) })

	e.MarkDone()
	assert.NotPanics(t, func() { PrintEpicWithIndex(0, e) })
}
