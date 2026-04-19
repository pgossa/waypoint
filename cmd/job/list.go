package job

import (
	"waypoint/model"
	"waypoint/ui"
)

func jobList(args []string) {
	showDone := len(args) == 1 && (args[0] == "--done" || args[0] == "-d")
	if len(args) > 1 || (len(args) == 1 && !showDone) {
		jobListUsage()
		return
	}

	jobs, err := getSortedJobs()
	if err != nil {
		ui.Error(err.Error())
		return
	}

	filtered := make([]*model.Job, 0)
	for _, job := range jobs {
		if !showDone && job.IsDone() {
			continue
		}
		filtered = append(filtered, job)
	}

	if len(filtered) == 0 {
		ui.Muted("No jobs, have a chill day!")
		return
	}

	items := make([]ui.ListItem, len(filtered))
	for i, job := range filtered {
		index := i + 1
		if showDone {
			index = 0 // no index in done view
		}
		items[i] = ui.ListItem{
			Index: index,
			Done:  job.IsDone(),
			Name:  job.GetName(),
		}
	}
	ui.List("JOB", items)
}

func jobListUsage() {
	ui.Info("Usage: wpt job list [--done|-d]")
	ui.Info("  Lists all pending jobs by default.")
	ui.Info("  Pass '--done' to include completed jobs.")
}
