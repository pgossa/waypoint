package cmd

import (
	"math/rand"
	"os"

	"waypoint/storage"
	"waypoint/ui"
)

func Next(args []string) {
	jobs, err := storage.GetJobs()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	tasks, err := storage.GetTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	all := make([]ui.NextItem, 0)

	for _, job := range jobs {
		if !job.IsDone() {
			all = append(all, ui.NextItem{Kind: "job", Name: job.GetName()})
		}
	}

	for _, task := range tasks {
		if !task.IsDone() {
			all = append(all, ui.NextItem{Kind: "task", Name: task.GetName(), Path: task.GetPath()})
		}
	}

	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })

	count := min(3, len(all))
	ui.NextUp(all[:count], len(all)-count)
}
