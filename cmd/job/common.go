package job

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
)

func getSortedJobs() ([]*model.Job, error) {
	jobs, err := storage.GetJobs()
	if err != nil {
		return nil, err
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].GetName() < jobs[j].GetName()
	})
	return jobs, nil
}

func getActiveJobs() ([]*model.Job, error) {
	jobs, err := getSortedJobs()
	if err != nil {
		return nil, err
	}
	active := make([]*model.Job, 0)
	for _, job := range jobs {
		if !job.IsDone() {
			active = append(active, job)
		}
	}
	return active, nil
}

func findJob(input string, jobs []*model.Job) (int, []int) {
	// try index first
	if n, err := strconv.Atoi(input); err == nil {
		if n >= 1 && n <= len(jobs) {
			return n - 1, []int{n - 1}
		}
		return -1, nil
	}
	matches := make([]int, 0)
	for i, job := range jobs {
		if strings.Contains(strings.ToLower(job.GetName()), strings.ToLower(input)) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 1 {
		return matches[0], matches
	}
	return -1, matches
}

func printJobWithIndex(i int, job *model.Job) {
	status := " "
	if job.IsDone() {
		status = "✓"
	}
	ui.Info(fmt.Sprintf("[%d] %s %s", i+1, status, job.GetName()))
}
