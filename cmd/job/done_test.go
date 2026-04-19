package job

import (
	"testing"

	"waypoint/model"

	"github.com/stretchr/testify/assert"
)

func makeJob(name string, done bool) *model.Job {
	job, _ := model.CreateJob(name)
	if done {
		job.MarkDone()
	}
	return job
}

// =============================================================================
// findJob index resolution (as used by jobDone to pick from getActiveJobs)
// =============================================================================

// jobDone calls getActiveJobs (pending-only, 1-based list) then findJob.
// Index "1" should resolve to position 0.
func TestJobDone_FindByIndex_FirstItem(t *testing.T) {
	jobs := []*model.Job{
		makeJob("build-api", false),
		makeJob("fix-tests", false),
	}
	idx, _ := findJob("1", jobs)
	assert.Equal(t, 0, idx)
}

// Index "2" resolves to position 1.
func TestJobDone_FindByIndex_SecondItem(t *testing.T) {
	jobs := []*model.Job{
		makeJob("build-api", false),
		makeJob("fix-tests", false),
	}
	idx, _ := findJob("2", jobs)
	assert.Equal(t, 1, idx)
}

// Index "0" is below the 1-based minimum → not found.
func TestJobDone_FindByIndex_Zero_NotFound(t *testing.T) {
	jobs := []*model.Job{makeJob("build-api", false)}
	idx, matches := findJob("0", jobs)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// Index beyond list length → not found.
func TestJobDone_FindByIndex_OutOfRange(t *testing.T) {
	jobs := []*model.Job{makeJob("only", false)}
	idx, matches := findJob("99", jobs)
	assert.Equal(t, -1, idx)
	assert.Nil(t, matches)
}

// =============================================================================
// findJob name resolution (as used by jobDone)
// =============================================================================

func TestJobDone_NoMatchFound(t *testing.T) {
	jobs := []*model.Job{
		makeJob("build-api", false),
		makeJob("fix-tests", false),
	}
	idx, matches := findJob("nonexistent", jobs)
	assert.Equal(t, -1, idx)
	assert.Empty(t, matches)
}

func TestJobDone_ExactMatchByName(t *testing.T) {
	jobs := []*model.Job{
		makeJob("build-api", false),
		makeJob("fix-tests", false),
	}
	idx, matches := findJob("build-api", jobs)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestJobDone_PartialMatch_SingleResult(t *testing.T) {
	jobs := []*model.Job{
		makeJob("build-api", false),
		makeJob("fix-tests", false),
	}
	idx, matches := findJob("build", jobs)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, matches)
}

func TestJobDone_PartialMatch_Ambiguous(t *testing.T) {
	jobs := []*model.Job{
		makeJob("build-api", false),
		makeJob("build-worker", false),
		makeJob("fix-tests", false),
	}
	idx, matches := findJob("build", jobs)
	assert.Equal(t, -1, idx)
	assert.Len(t, matches, 2)
}

// =============================================================================
// IsDone / MarkDone state
// =============================================================================

func TestJobDone_MarkDone_Transition(t *testing.T) {
	job := makeJob("pending-job", false)
	assert.False(t, job.IsDone())
	job.MarkDone()
	assert.True(t, job.IsDone())
}

func TestJobDone_MarkDone_Idempotent(t *testing.T) {
	job := makeJob("idempotent-job", false)
	job.MarkDone()
	job.MarkDone()
	assert.True(t, job.IsDone())
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestJobDoneUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { jobDoneUsage() })
}
