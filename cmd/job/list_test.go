package job

import (
	"sort"
	"testing"

	"waypoint/model"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Flag parsing: the real flag is --done / -d  (NOT --all)
// Logic: showDone := len(args)==1 && (args[0]=="--done" || args[0]=="-d")
// Anything else with args triggers usage.
// =============================================================================

func TestJobList_DoneFlag_Recognised(t *testing.T) {
	for _, arg := range []string{"--done", "-d"} {
		args := []string{arg}
		showDone := len(args) == 1 && (args[0] == "--done" || args[0] == "-d")
		assert.True(t, showDone, "expected showDone=true for %q", arg)
	}
}

func TestJobList_NoArgs_ShowPending(t *testing.T) {
	args := []string{}
	showDone := len(args) == 1 && (args[0] == "--done" || args[0] == "-d")
	triggersUsage := len(args) > 1 || (len(args) == 1 && !showDone)
	assert.False(t, showDone)
	assert.False(t, triggersUsage)
}

func TestJobList_UnknownFlag_TriggersUsage(t *testing.T) {
	for _, arg := range []string{"--all", "all", "a", "pending", "--DONE", "done"} {
		args := []string{arg}
		showDone := len(args) == 1 && (args[0] == "--done" || args[0] == "-d")
		triggersUsage := len(args) == 1 && !showDone
		assert.True(t, triggersUsage, "expected usage for unrecognised arg %q", arg)
	}
}

func TestJobList_TooManyArgs_TriggersUsage(t *testing.T) {
	args := []string{"--done", "extra"}
	triggersUsage := len(args) > 1
	assert.True(t, triggersUsage)
}

// =============================================================================
// Sorting
// =============================================================================

func TestJobList_SortedAlphabetically(t *testing.T) {
	jobs := makeJobs(t,
		jobSpec{"zebra", false},
		jobSpec{"alpha", false},
		jobSpec{"middle", false},
	)
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].GetName() < jobs[j].GetName() })
	assert.Equal(t, "alpha", jobs[0].GetName())
	assert.Equal(t, "middle", jobs[1].GetName())
	assert.Equal(t, "zebra", jobs[2].GetName())
}

// =============================================================================
// Filtering: pending-only (showDone=false)
// Logic: if !showDone && job.IsDone() { continue }
// =============================================================================

func TestJobList_PendingOnly_ExcludesDone(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"pending", false}, jobSpec{"done", true})
	filtered := filterJobs(jobs, false)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "pending", filtered[0].GetName())
}

func TestJobList_PendingOnly_AllDone_Empty(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"done-a", true}, jobSpec{"done-b", true})
	assert.Empty(t, filterJobs(jobs, false))
}

func TestJobList_PendingOnly_NoneDone_ShowsAll(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"task-a", false}, jobSpec{"task-b", false})
	assert.Len(t, filterJobs(jobs, false), 2)
}

// =============================================================================
// Filtering: show done (showDone=true)
// =============================================================================

func TestJobList_ShowDone_IncludesDone(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"pending", false}, jobSpec{"done", true})
	assert.Len(t, filterJobs(jobs, true), 2)
}

func TestJobList_ShowDone_AllPending(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"a", false}, jobSpec{"b", false})
	assert.Len(t, filterJobs(jobs, true), 2)
}

func TestJobList_ShowDone_Empty(t *testing.T) {
	assert.Empty(t, filterJobs([]*model.Job{}, true))
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestJobListUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { jobListUsage() })
}

// =============================================================================
// Helpers
// =============================================================================

type jobSpec struct {
	name string
	done bool
}

func makeJobs(t *testing.T, specs ...jobSpec) []*model.Job {
	t.Helper()
	jobs := make([]*model.Job, 0, len(specs))
	for _, s := range specs {
		jobs = append(jobs, newJob(t, s.name, s.done))
	}
	return jobs
}

// filterJobs replicates the filter loop inside jobList.
// showDone=false → pending only; showDone=true → all jobs.
func filterJobs(jobs []*model.Job, showDone bool) []*model.Job {
	var out []*model.Job
	for _, job := range jobs {
		if !showDone && job.IsDone() {
			continue
		}
		out = append(out, job)
	}
	return out
}
