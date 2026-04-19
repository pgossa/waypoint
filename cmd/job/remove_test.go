package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// findJob resolution (drives jobRemove branching)
// =============================================================================

func TestJobRemove_NoMatchFound(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"build-api", false}, jobSpec{"fix-tests", false})
	idx, matches := findJob("nonexistent", jobs)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

func TestJobRemove_SingleMatchByName(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"build-api", false}, jobSpec{"fix-tests", false})
	idx, matches := findJob("build-api", jobs)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestJobRemove_SingleMatchByPartialName(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"build-api", false}, jobSpec{"fix-tests", false})
	idx, matches := findJob("fix", jobs)
	assert.Equal(t, 1, idx)
	assert.Equal(t, []int{1}, matches)
}

func TestJobRemove_MultipleMatches(t *testing.T) {
	jobs := makeJobs(t,
		jobSpec{"build-api", false},
		jobSpec{"build-worker", false},
		jobSpec{"fix-tests", false},
	)
	idx, matches := findJob("build", jobs)
	assert.Equal(t, -1, idx)
	assert.ElementsMatch(t, []int{0, 1}, matches)
}

// 1-based indexing: "2" resolves to position 1.
func TestJobRemove_MatchByIndex(t *testing.T) {
	jobs := makeJobs(t,
		jobSpec{"alpha", false},
		jobSpec{"beta", false},
		jobSpec{"gamma", false},
	)
	idx, matches := findJob("2", jobs)
	assert.Equal(t, 1, idx)
	assert.Equal(t, []int{1}, matches)
}

func TestJobRemove_IndexOutOfRange(t *testing.T) {
	jobs := makeJobs(t, jobSpec{"only", false})
	idx, matches := findJob("99", jobs)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// =============================================================================
// Job identity — GetID is used by storage.RemoveJob
// =============================================================================

func TestJobRemove_JobHasStableID(t *testing.T) {
	job := newJob(t, "removable", false)
	assert.Equal(t, job.GetID(), job.GetID())
	assert.NotEmpty(t, job.GetID())
}

func TestJobRemove_DifferentJobs_DifferentIDs(t *testing.T) {
	a := newJob(t, "job-one", false)
	b := newJob(t, "job-two", false)
	assert.NotEqual(t, a.GetID(), b.GetID())
}

// jobRemove has no IsDone guard — both pending and done jobs can be removed.
func TestJobRemove_CanRemovePendingOrDone(t *testing.T) {
	pending := newJob(t, "pending", false)
	done := newJob(t, "done", true)
	assert.NotEmpty(t, pending.GetID())
	assert.NotEmpty(t, done.GetID())
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestJobRemoveUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { jobRemoveUsage() })
}
