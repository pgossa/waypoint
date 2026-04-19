package job

import (
	"os"

	"waypoint/storage"
	"waypoint/ui"
)

func jobDone(args []string) {
	if len(args) != 1 {
		jobDoneUsage()
		os.Exit(1)
	}

	jobs, err := getActiveJobs()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	idx, matches := findJob(args[0], jobs)

	if idx == -1 && len(matches) == 0 {
		ui.Error("no job found matching: " + args[0])
		os.Exit(1)
	}

	if len(matches) > 1 {
		ui.Error("multiple jobs match, use the index instead:")
		for _, i := range matches {
			printJobWithIndex(i, jobs[i])
		}
		os.Exit(1)
	}

	job := jobs[idx]
	if job.IsDone() {
		ui.Error("job is already done: " + job.GetName())
		os.Exit(1)
	}

	job.MarkDone()
	if err := storage.SaveJob(job); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("job done: " + job.GetName())
}

func jobDoneUsage() {
	ui.Info("Usage: wpt job done <index|name>")
	ui.Info("  <index>  The job index from 'wpt job list'")
	ui.Info("  <name>   Full or partial job name")
}
