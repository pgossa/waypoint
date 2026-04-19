package task

import (
	"testing"

	"waypoint/config"
	"waypoint/model"
	"waypoint/storage"

	"github.com/stretchr/testify/assert"
)

func setupTaskStorage(t *testing.T) {
	t.Helper()
	config.Reset()
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(config.Reset)
}

func newTask(t *testing.T, name string, path []string) *model.Task {
	t.Helper()
	task, err := model.CreateTask(name, path)
	assert.NoError(t, err)
	return task
}

// =============================================================================
// findTask — name-only search (no numeric index support unlike findJob)
// =============================================================================

// Single exact match: returns (idx, []int{idx}).
func TestFindTask_ExactNameMatch(t *testing.T) {
	tasks := []*model.Task{
		newTask(t, "build-api", []string{"root"}),
		newTask(t, "fix-tests", []string{"root"}),
	}
	idx, matches := findTask("build-api", tasks)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

// Unambiguous partial match.
func TestFindTask_PartialName_SingleMatch(t *testing.T) {
	tasks := []*model.Task{
		newTask(t, "build-api", []string{"root"}),
		newTask(t, "fix-tests", []string{"root"}),
	}
	idx, matches := findTask("fix", tasks)
	assert.Equal(t, 1, idx)
	assert.Equal(t, []int{1}, matches)
}

// Ambiguous partial match: returns (-1, all indices).
func TestFindTask_PartialName_MultipleMatches(t *testing.T) {
	tasks := []*model.Task{
		newTask(t, "build-api", []string{"root"}),
		newTask(t, "build-worker", []string{"root"}),
		newTask(t, "fix-tests", []string{"root"}),
	}
	idx, matches := findTask("build", tasks)
	assert.Equal(t, -1, idx)
	assert.ElementsMatch(t, []int{0, 1}, matches)
}

// No match: returns (-1, empty non-nil slice).
func TestFindTask_NoMatch(t *testing.T) {
	tasks := []*model.Task{
		newTask(t, "build-api", []string{"root"}),
	}
	idx, matches := findTask("nonexistent", tasks)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
	assert.NotNil(t, matches)
}

// Numeric strings are NOT treated as indices for tasks — they are name-searched.
func TestFindTask_NumericInput_TreatedAsName(t *testing.T) {
	tasks := []*model.Task{
		newTask(t, "build-api", []string{"root"}),
		newTask(t, "fix-tests", []string{"root"}),
	}
	// "1" does not match any task name → empty
	idx, matches := findTask("1", tasks)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

func TestFindTask_CaseInsensitive(t *testing.T) {
	tasks := []*model.Task{newTask(t, "Build-API", []string{"root"})}

	idx, _ := findTask("build-api", tasks)
	assert.Equal(t, 0, idx)

	idx, _ = findTask("BUILD-API", tasks)
	assert.Equal(t, 0, idx)
}

// Empty string matches every task name.
func TestFindTask_EmptyInput_MatchesAll(t *testing.T) {
	tasks := []*model.Task{
		newTask(t, "alpha", []string{"root"}),
		newTask(t, "beta", []string{"root"}),
	}
	idx, matches := findTask("", tasks)
	assert.Equal(t, -1, idx)
	assert.Len(t, matches, 2)
}

func TestFindTask_EmptyList(t *testing.T) {
	idx, matches := findTask("anything", []*model.Task{})
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

// =============================================================================
// pathHasPrefix — task path starts with given prefix
// =============================================================================

func TestPathHasPrefix_ExactMatch(t *testing.T) {
	assert.True(t, pathHasPrefix([]string{"home", "user"}, []string{"home", "user"}))
}

func TestPathHasPrefix_TaskLonger_Matches(t *testing.T) {
	// task is deeper than prefix — should match
	assert.True(t, pathHasPrefix([]string{"home", "user", "projects"}, []string{"home", "user"}))
}

func TestPathHasPrefix_TaskShorterThanPrefix_NoMatch(t *testing.T) {
	assert.False(t, pathHasPrefix([]string{"home"}, []string{"home", "user"}))
}

func TestPathHasPrefix_DifferentContent_NoMatch(t *testing.T) {
	assert.False(t, pathHasPrefix([]string{"work", "user"}, []string{"home", "user"}))
}

func TestPathHasPrefix_EmptyPrefix_AlwaysTrue(t *testing.T) {
	// Empty prefix is a prefix of everything.
	assert.True(t, pathHasPrefix([]string{"home", "user"}, []string{}))
	assert.True(t, pathHasPrefix([]string{}, []string{}))
}

func TestPathHasPrefix_BothEmpty(t *testing.T) {
	assert.True(t, pathHasPrefix([]string{}, []string{}))
}

func TestPathHasPrefix_CaseSensitive(t *testing.T) {
	assert.False(t, pathHasPrefix([]string{"Home"}, []string{"home"}))
}

func TestPathHasPrefix_DiffersAtSecondSegment(t *testing.T) {
	assert.False(t, pathHasPrefix([]string{"home", "alice"}, []string{"home", "bob"}))
}

// =============================================================================
// getSortedTasks / getActiveTasks / printTaskWithIndex — storage-integrated
// =============================================================================

func TestGetSortedTasks_ReturnsSorted(t *testing.T) {
	setupTaskStorage(t)

	for _, name := range []string{"gamma", "alpha", "beta"} {
		task, _ := model.CreateTask(name, []string{"root"})
		assert.NoError(t, storage.SaveTask(task))
	}

	tasks, err := getSortedTasks()
	assert.NoError(t, err)
	assert.Len(t, tasks, 3)
	assert.Equal(t, "alpha", tasks[0].GetName())
	assert.Equal(t, "beta", tasks[1].GetName())
	assert.Equal(t, "gamma", tasks[2].GetName())
}

func TestGetSortedTasks_Empty(t *testing.T) {
	setupTaskStorage(t)
	tasks, err := getSortedTasks()
	assert.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestGetActiveTasks_FiltersOutDone(t *testing.T) {
	setupTaskStorage(t)

	active, _ := model.CreateTask("active", []string{"root"})
	done, _ := model.CreateTask("done", []string{"root"})
	done.MarkDone()
	assert.NoError(t, storage.SaveTask(active))
	assert.NoError(t, storage.SaveTask(done))

	result, err := getActiveTasks()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "active", result[0].GetName())
}

func TestGetActiveTasks_AllDone_ReturnsEmpty(t *testing.T) {
	setupTaskStorage(t)

	task, _ := model.CreateTask("done", []string{"root"})
	task.MarkDone()
	assert.NoError(t, storage.SaveTask(task))

	result, err := getActiveTasks()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestPrintTaskWithIndex_NoPanic(t *testing.T) {
	pending := newTask(t, "pending-task", []string{"root"})
	done := newTask(t, "done-task", []string{"root"})
	done.MarkDone()

	assert.NotPanics(t, func() { printTaskWithIndex(0, pending) })
	assert.NotPanics(t, func() { printTaskWithIndex(0, done) })
	assert.NotPanics(t, func() { printTaskWithIndex(99, pending) })
}
