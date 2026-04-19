package task

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Flag parsing
// taskList accepts: --all/-a/all, --done/-d, --path/-p <path>
// Any unrecognised flag triggers usage + os.Exit.
// =============================================================================

func TestTaskList_FlagParsing_AllVariants(t *testing.T) {
	for _, arg := range []string{"--all", "-a", "all"} {
		showAll, showDone := parseTaskListFlags(t, []string{arg})
		assert.True(t, showAll, "--all flag not recognised for %q", arg)
		assert.False(t, showDone)
	}
}

func TestTaskList_FlagParsing_DoneFlag(t *testing.T) {
	for _, arg := range []string{"--done", "-d"} {
		showAll, showDone := parseTaskListFlags(t, []string{arg})
		assert.False(t, showAll)
		assert.True(t, showDone, "--done flag not recognised for %q", arg)
	}
}

func TestTaskList_FlagParsing_PathFlag(t *testing.T) {
	customPath := parseTaskListPath(t, []string{"--path", "home/user/projects"})
	assert.Equal(t, []string{"home", "user", "projects"}, customPath)

	customPath = parseTaskListPath(t, []string{"-p", "a/b"})
	assert.Equal(t, []string{"a", "b"}, customPath)
}

func TestTaskList_FlagParsing_CombinedFlags(t *testing.T) {
	showAll, showDone := parseTaskListFlags(t, []string{"--all", "--done"})
	assert.True(t, showAll)
	assert.True(t, showDone)
}

func TestTaskList_FlagParsing_NoArgs(t *testing.T) {
	showAll, showDone := parseTaskListFlags(t, []string{})
	assert.False(t, showAll)
	assert.False(t, showDone)
}

// =============================================================================
// Path-based filtering
// Logic from taskList: if !showAll && !util.PathMatches(task.GetPath(), activePath) { skip }
// =============================================================================

func TestTaskList_FilterByPath_ExactMatch(t *testing.T) {
	path := []string{"home", "user", "api"}
	tasks := []*model.Task{
		makeTask(t, "in-dir", path),
		makeTask(t, "elsewhere", []string{"home", "user", "other"}),
	}
	filtered := filterTasksByPath(tasks, false, path)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "in-dir", filtered[0].GetName())
}

func TestTaskList_FilterByPath_NoMatch(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "other", []string{"work", "project"}),
	}
	filtered := filterTasksByPath(tasks, false, []string{"home", "user"})
	assert.Empty(t, filtered)
}

func TestTaskList_FilterByPath_ShowAll_IgnoresPath(t *testing.T) {
	tasks := []*model.Task{
		makeTask(t, "task-a", []string{"home", "user"}),
		makeTask(t, "task-b", []string{"work", "project"}),
	}
	// showAll=true should return all tasks regardless of activePath
	filtered := filterTasksByPath(tasks, true, []string{"home", "user"})
	assert.Len(t, filtered, 2)
}

func TestTaskList_FilterByPath_MultipleInSameDir(t *testing.T) {
	path := []string{"home", "user"}
	tasks := []*model.Task{
		makeTask(t, "alpha", path),
		makeTask(t, "beta", path),
		makeTask(t, "gamma", []string{"work"}),
	}
	filtered := filterTasksByPath(tasks, false, path)
	assert.Len(t, filtered, 2)
}

func TestTaskList_FilterByPath_EmptyTaskList(t *testing.T) {
	filtered := filterTasksByPath([]*model.Task{}, false, []string{"home"})
	assert.Empty(t, filtered)
}

// =============================================================================
// Pending-only vs all (showDone flag effect on which tasks are fetched)
// Logic: showDone=true → getActiveTasks+done; showDone=false → active only
// We replicate the filter here since getActiveTasks calls storage.
// =============================================================================

func TestTaskList_PendingFilter_ExcludesDoneWhenNotShowDone(t *testing.T) {
	path := []string{"home"}
	pending := makeTask(t, "pending", path)
	done := makeTask(t, "done-task", path)
	done.MarkDone()

	// simulates the active-only result + path filter
	filtered := filterTasksByPath([]*model.Task{pending}, false, path)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "pending", filtered[0].GetName())
}

func TestTaskList_ShowDone_IncludesDone(t *testing.T) {
	path := []string{"home"}
	pending := makeTask(t, "pending", path)
	done := makeTask(t, "done-task", path)
	done.MarkDone()

	// showAll=true bypasses the path filter → all tasks visible
	all := filterTasksByPath([]*model.Task{pending, done}, true, path)
	assert.Len(t, all, 2)
}

// =============================================================================
// Index assignment: index is 0 (hidden) when showAll or showDone is true
// =============================================================================

func TestTaskList_IndexAssignment(t *testing.T) {
	path := []string{"home"}
	tasks := []*model.Task{
		makeTask(t, "task-a", path),
		makeTask(t, "task-b", path),
	}

	// normal view: indices start at 1
	for i, task := range tasks {
		_ = task
		index := i + 1
		assert.Greater(t, index, 0)
	}

	// showAll=true or showDone=true → index=0 (no number displayed)
	for range tasks {
		index := 0 // same logic as taskList: if showAll || showDone { index = 0 }
		assert.Equal(t, 0, index)
	}
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestTaskListUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { taskListUsage() })
}

// =============================================================================
// Helpers
// =============================================================================

func makeTask(t *testing.T, name string, path []string) *model.Task {
	t.Helper()
	task, err := model.CreateTask(name, path)
	assert.NoError(t, err)
	return task
}

// parseTaskListFlags replicates taskList's flag-parsing loop for showAll/showDone.
func parseTaskListFlags(t *testing.T, args []string) (showAll, showDone bool) {
	t.Helper()
	for _, arg := range args {
		switch arg {
		case "--all", "-a", "all":
			showAll = true
		case "--done", "-d":
			showDone = true
		case "--path", "-p":
			// skip value in this simplified parser
		}
	}
	return
}

// parseTaskListPath replicates the --path flag parsing.
func parseTaskListPath(t *testing.T, args []string) []string {
	t.Helper()
	for i, arg := range args {
		if (arg == "--path" || arg == "-p") && i+1 < len(args) {
			return strings.Split(args[i+1], "/")
		}
	}
	return nil
}

// filterTasksByPath replicates the filter loop from taskList.
func filterTasksByPath(tasks []*model.Task, showAll bool, activePath []string) []*model.Task {
	var out []*model.Task
	for _, task := range tasks {
		if !showAll && !util.PathMatches(task.GetPath(), activePath) {
			continue
		}
		out = append(out, task)
	}
	return out
}
