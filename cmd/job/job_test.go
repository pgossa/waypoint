package job

import (
	"sort"
	"testing"

	"waypoint/config"
	"waypoint/model"
	"waypoint/storage"

	"github.com/stretchr/testify/assert"
)

func setupJobStorage(t *testing.T) {
	t.Helper()
	config.Reset()
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(config.Reset)
}

func newJob(t *testing.T, name string, done bool) *model.Job {
	t.Helper()
	job, err := model.CreateJob(name)
	assert.NoError(t, err)
	if done {
		job.MarkDone()
	}
	return job
}

// =============================================================================
// getSortedJobs / getActiveJobs — storage-integrated
// =============================================================================

func TestGetSortedJobs_ReturnsSorted(t *testing.T) {
	setupJobStorage(t)

	for _, name := range []string{"gamma", "alpha", "beta"} {
		j, _ := model.CreateJob(name)
		assert.NoError(t, storage.SaveJob(j))
	}

	jobs, err := getSortedJobs()
	assert.NoError(t, err)
	assert.Len(t, jobs, 3)
	assert.Equal(t, "alpha", jobs[0].GetName())
	assert.Equal(t, "beta", jobs[1].GetName())
	assert.Equal(t, "gamma", jobs[2].GetName())
}

func TestGetSortedJobs_Empty(t *testing.T) {
	setupJobStorage(t)
	jobs, err := getSortedJobs()
	assert.NoError(t, err)
	assert.Empty(t, jobs)
}

func TestGetActiveJobs_FiltersOutDone(t *testing.T) {
	setupJobStorage(t)

	active, _ := model.CreateJob("active")
	done, _ := model.CreateJob("done")
	done.MarkDone()
	assert.NoError(t, storage.SaveJob(active))
	assert.NoError(t, storage.SaveJob(done))

	result, err := getActiveJobs()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "active", result[0].GetName())
}

func TestGetActiveJobs_AllDone_ReturnsEmpty(t *testing.T) {
	setupJobStorage(t)

	j, _ := model.CreateJob("done")
	j.MarkDone()
	assert.NoError(t, storage.SaveJob(j))

	result, err := getActiveJobs()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// =============================================================================
// getSortedJobs() – sorting logic (tested via sortJobs helper)
// =============================================================================

func TestSortJobs_AlphabeticalOrder(t *testing.T) {
	jobs := []*model.Job{
		newJob(t, "zebra-task", false),
		newJob(t, "alpha-task", false),
		newJob(t, "middle-task", false),
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].GetName() < jobs[j].GetName() })

	assert.Equal(t, "alpha-task", jobs[0].GetName())
	assert.Equal(t, "middle-task", jobs[1].GetName())
	assert.Equal(t, "zebra-task", jobs[2].GetName())
}

func TestSortJobs_AlreadySorted(t *testing.T) {
	jobs := []*model.Job{newJob(t, "aaa", false), newJob(t, "bbb", false), newJob(t, "ccc", false)}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].GetName() < jobs[j].GetName() })
	assert.Equal(t, "aaa", jobs[0].GetName())
	assert.Equal(t, "ccc", jobs[2].GetName())
}

func TestSortJobs_Empty(t *testing.T) {
	var jobs []*model.Job
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].GetName() < jobs[j].GetName() })
	assert.Empty(t, jobs)
}

// =============================================================================
// findJob() – numeric index resolution
// =============================================================================

// findJob uses 1-based indexing. In-range returns (n-1, []int{n-1}).
func TestFindJob_IndexInRange(t *testing.T) {
	jobs := []*model.Job{
		newJob(t, "alpha", false),
		newJob(t, "beta", false),
		newJob(t, "gamma", false),
	}

	idx, matches := findJob("1", jobs)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)

	idx, matches = findJob("3", jobs)
	assert.Equal(t, 2, idx)
	assert.Equal(t, []int{2}, matches)
}

// Index 0 is below the 1-based minimum: returns (-1, nil).
func TestFindJob_IndexZero_IsOutOfRange(t *testing.T) {
	jobs := []*model.Job{newJob(t, "alpha", false)}
	idx, matches := findJob("0", jobs)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// Index beyond the list length: returns (-1, nil).
func TestFindJob_IndexOutOfRange(t *testing.T) {
	jobs := []*model.Job{newJob(t, "only", false)}
	idx, matches := findJob("99", jobs)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// Negative numbers are parsed by Atoi but fail the >=1 guard.
func TestFindJob_NegativeIndex(t *testing.T) {
	jobs := []*model.Job{newJob(t, "alpha", false)}
	idx, matches := findJob("-1", jobs)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// =============================================================================
// findJob() – name resolution
// =============================================================================

// Single exact match returns the job index with a 1-element matches slice.
func TestFindJob_ExactNameMatch(t *testing.T) {
	jobs := []*model.Job{
		newJob(t, "build-api", false),
		newJob(t, "fix-tests", false),
	}
	idx, matches := findJob("build-api", jobs)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

// Unambiguous partial match behaves the same as an exact match.
func TestFindJob_PartialName_SingleMatch(t *testing.T) {
	jobs := []*model.Job{
		newJob(t, "build-api", false),
		newJob(t, "fix-tests", false),
	}
	idx, matches := findJob("fix", jobs)
	assert.Equal(t, 1, idx)
	assert.Equal(t, []int{1}, matches)
}

// Ambiguous partial match returns (-1, allMatchingIndices).
func TestFindJob_PartialName_MultipleMatches(t *testing.T) {
	jobs := []*model.Job{
		newJob(t, "build-api", false),
		newJob(t, "build-worker", false),
		newJob(t, "fix-tests", false),
	}
	idx, matches := findJob("build", jobs)
	assert.Equal(t, -1, idx)
	assert.ElementsMatch(t, []int{0, 1}, matches)
}

// No name match returns (-1, empty-non-nil slice).
func TestFindJob_NoMatch(t *testing.T) {
	jobs := []*model.Job{
		newJob(t, "build-api", false),
		newJob(t, "fix-tests", false),
	}
	idx, matches := findJob("nonexistent", jobs)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
	assert.NotNil(t, matches) // empty slice, not nil
}

func TestFindJob_CaseInsensitive(t *testing.T) {
	jobs := []*model.Job{newJob(t, "Build-API", false)}

	idx, _ := findJob("build-api", jobs)
	assert.Equal(t, 0, idx)

	idx, _ = findJob("BUILD-API", jobs)
	assert.Equal(t, 0, idx)

	idx, _ = findJob("bUiLd", jobs)
	assert.Equal(t, 0, idx)
}

// Empty string matches every job (strings.Contains("x", "") is always true).
func TestFindJob_EmptyInput_MatchesAll(t *testing.T) {
	jobs := []*model.Job{newJob(t, "alpha", false), newJob(t, "beta", false)}
	idx, matches := findJob("", jobs)
	assert.Equal(t, -1, idx)
	assert.Len(t, matches, 2)
}

func TestFindJob_EmptyJobList(t *testing.T) {
	idx, matches := findJob("anything", []*model.Job{})
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

// =============================================================================
// printJobWithIndex() – smoke test (no panic)
// =============================================================================

func TestPrintJobWithIndex_NoPanic(t *testing.T) {
	pending := newJob(t, "pending-job", false)
	done := newJob(t, "done-job", true)

	assert.NotPanics(t, func() { printJobWithIndex(0, pending) })
	assert.NotPanics(t, func() { printJobWithIndex(0, done) })
	assert.NotPanics(t, func() { printJobWithIndex(99, pending) })
}
