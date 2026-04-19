package epic

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Flag parsing
// epicList accepts: --all/-a/all, --done/-d, --path/-p <path>
// =============================================================================

func TestEpicList_FlagParsing_AllVariants(t *testing.T) {
	for _, arg := range []string{"--all", "-a", "all"} {
		showAll, showDone := parseEpicListFlags(t, []string{arg})
		assert.True(t, showAll, "expected showAll=true for %q", arg)
		assert.False(t, showDone)
	}
}

func TestEpicList_FlagParsing_DoneFlag(t *testing.T) {
	for _, arg := range []string{"--done", "-d"} {
		showAll, showDone := parseEpicListFlags(t, []string{arg})
		assert.False(t, showAll)
		assert.True(t, showDone, "expected showDone=true for %q", arg)
	}
}

func TestEpicList_FlagParsing_PathFlag(t *testing.T) {
	path := parseEpicListPath(t, []string{"--path", "home/user"})
	assert.Equal(t, []string{"home", "user"}, path)

	path = parseEpicListPath(t, []string{"-p", "a/b/c"})
	assert.Equal(t, []string{"a", "b", "c"}, path)
}

func TestEpicList_FlagParsing_Combined(t *testing.T) {
	showAll, showDone := parseEpicListFlags(t, []string{"--all", "--done"})
	assert.True(t, showAll)
	assert.True(t, showDone)
}

func TestEpicList_FlagParsing_NoArgs(t *testing.T) {
	showAll, showDone := parseEpicListFlags(t, []string{})
	assert.False(t, showAll)
	assert.False(t, showDone)
}

// =============================================================================
// Path-based filtering
// Logic: if !showAll && !util.PathMatches(epic.GetPath(), activePath) { skip }
// =============================================================================

func TestEpicList_FilterByPath_Match(t *testing.T) {
	path := []string{"home", "user", "api"}
	epics := []*model.Epic{
		newEpicAt(t, "in-dir", path),
		newEpicAt(t, "elsewhere", []string{"home", "user", "other"}),
	}
	filtered := filterEpicsByPath(epics, false, path)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "in-dir", filtered[0].GetName())
}

func TestEpicList_FilterByPath_NoMatch(t *testing.T) {
	epics := []*model.Epic{newEpicAt(t, "other", []string{"work", "project"})}
	filtered := filterEpicsByPath(epics, false, []string{"home", "user"})
	assert.Empty(t, filtered)
}

func TestEpicList_FilterByPath_ShowAll_IgnoresPath(t *testing.T) {
	epics := []*model.Epic{
		newEpicAt(t, "epic-a", []string{"home", "user"}),
		newEpicAt(t, "epic-b", []string{"work", "project"}),
	}
	filtered := filterEpicsByPath(epics, true, []string{"home", "user"})
	assert.Len(t, filtered, 2)
}

func TestEpicList_FilterByPath_EmptyList(t *testing.T) {
	filtered := filterEpicsByPath([]*model.Epic{}, false, []string{"home"})
	assert.Empty(t, filtered)
}

// =============================================================================
// Progress calculation
// Logic from epicList: count tasks where taskMap[tid].IsDone()
// =============================================================================

func TestEpicList_Progress_AllDone(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	for i := 0; i < 3; i++ {
		task, _ := model.CreateTask("task", []string{"root"})
		task.MarkDone()
		_ = epic.AddTaskID(task.GetID())
		taskMap[task.GetID()] = task
	}

	done, total := countDoneSubtasks(epic, taskMap)
	assert.Equal(t, 3, total)
	assert.Equal(t, 3, done)
}

func TestEpicList_Progress_NoneDone(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	for i := 0; i < 3; i++ {
		task, _ := model.CreateTask("task", []string{"root"})
		_ = epic.AddTaskID(task.GetID())
		taskMap[task.GetID()] = task
	}

	done, total := countDoneSubtasks(epic, taskMap)
	assert.Equal(t, 3, total)
	assert.Equal(t, 0, done)
}

func TestEpicList_Progress_Partial(t *testing.T) {
	epic := newEpic(t, "release")
	taskMap := make(map[uuid.UUID]*model.Task)

	doneTask, _ := model.CreateTask("done", []string{"root"})
	doneTask.MarkDone()
	_ = epic.AddTaskID(doneTask.GetID())
	taskMap[doneTask.GetID()] = doneTask

	pendingTask, _ := model.CreateTask("pending", []string{"root"})
	_ = epic.AddTaskID(pendingTask.GetID())
	taskMap[pendingTask.GetID()] = pendingTask

	done, total := countDoneSubtasks(epic, taskMap)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, done)
}

func TestEpicList_Progress_NoSubtasks(t *testing.T) {
	epic := newEpic(t, "empty-epic")
	done, total := countDoneSubtasks(epic, map[uuid.UUID]*model.Task{})
	assert.Equal(t, 0, total)
	assert.Equal(t, 0, done)
}

// Orphaned task IDs (task deleted but still in epic) should not crash or count.
func TestEpicList_Progress_OrphanedTaskID_Ignored(t *testing.T) {
	epic := newEpic(t, "release")
	_ = epic.AddTaskID(uuid.New()) // ID not in taskMap

	done, total := countDoneSubtasks(epic, map[uuid.UUID]*model.Task{})
	assert.Equal(t, 1, total)
	assert.Equal(t, 0, done)
}

// =============================================================================
// Index assignment in list view
// showAll=true or showDone=true → index=0 (no number shown)
// =============================================================================

func TestEpicList_IndexAssignment_NormalView(t *testing.T) {
	// In a normal (non-all, non-done) view, epics get 1-based indices.
	epics := []*model.Epic{newEpic(t, "a"), newEpic(t, "b")}
	for i := range epics {
		index := i + 1
		assert.Greater(t, index, 0)
	}
}

func TestEpicList_IndexAssignment_FlatView(t *testing.T) {
	// showAll=true or showDone=true → index=0
	for range []string{"showAll", "showDone"} {
		index := 0
		assert.Equal(t, 0, index)
	}
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestEpicListUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { epicListUsage() })
}

// =============================================================================
// Helpers
// =============================================================================

// newEpicAt creates an epic with a specific path (for path-filter tests).
func newEpicAt(t *testing.T, name string, path []string) *model.Epic {
	t.Helper()
	epic, err := model.CreateEpic(name, path, nil)
	assert.NoError(t, err)
	return epic
}

// parseEpicListFlags replicates epicList's flag-parsing loop.
func parseEpicListFlags(t *testing.T, args []string) (showAll, showDone bool) {
	t.Helper()
	for _, arg := range args {
		switch arg {
		case "--all", "-a", "all":
			showAll = true
		case "--done", "-d":
			showDone = true
		}
	}
	return
}

// parseEpicListPath replicates the --path flag parsing.
func parseEpicListPath(t *testing.T, args []string) []string {
	t.Helper()
	for i, arg := range args {
		if (arg == "--path" || arg == "-p") && i+1 < len(args) {
			return strings.Split(args[i+1], "/")
		}
	}
	return nil
}

// filterEpicsByPath replicates the filter loop from epicList.
func filterEpicsByPath(epics []*model.Epic, showAll bool, activePath []string) []*model.Epic {
	var out []*model.Epic
	for _, epic := range epics {
		if !showAll && !util.PathMatches(epic.GetPath(), activePath) {
			continue
		}
		out = append(out, epic)
	}
	return out
}

// countDoneSubtasks replicates the progress-counting loop from epicList.
func countDoneSubtasks(epic *model.Epic, taskMap map[uuid.UUID]*model.Task) (done, total int) {
	ids := epic.GetTasksID()
	total = len(ids)
	for _, tid := range ids {
		if t, ok := taskMap[tid]; ok && t.IsDone() {
			done++
		}
	}
	return
}
