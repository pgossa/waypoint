package cmd

import (
	"time"

	"waypoint/config"
	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
	"waypoint/util"
)

// List shows all tasks and jobs regardless of path.
func List(args []string) {
	showDone := false
	for _, a := range args {
		if a == "--done" || a == "-d" {
			showDone = true
		}
	}

	tasks, err := storage.GetTasks()
	if err != nil {
		ui.Error(err.Error())
		return
	}

	var taskItems []ui.ListItem
	for _, task := range tasks {
		if !showDone && task.IsDone() {
			continue
		}
		sub := ""
		if task.GetEpicID() != nil {
			sub = "part of an epic"
		}
		taskItems = append(taskItems, ui.ListItem{
			Done: task.IsDone(),
			Name: task.GetName(),
			Path: pathJoin(task.GetPath()),
			Sub:  sub,
		})
	}
	if len(taskItems) > 0 {
		ui.List("TASK", taskItems)
	}

	jobs, err := storage.GetJobs()
	if err != nil {
		ui.Error(err.Error())
		return
	}

	var jobItems []ui.ListItem
	for _, job := range jobs {
		if !showDone && job.IsDone() {
			continue
		}
		jobItems = append(jobItems, ui.ListItem{
			Done: job.IsDone(),
			Name: job.GetName(),
		})
	}
	if len(jobItems) > 0 {
		ui.List("JOB", jobItems)
	}
}

// Cd is the shell hook called on directory change — shows only items
// matching the current path, and rate-limits jobs to once per 15 minutes.
func Cd(args []string) {
	currentPath := util.GetPathSplit()

	tasks, err := storage.GetTasks()
	if err != nil {
		ui.Error(err.Error())
		return
	}

	matching := make([]*model.Task, 0)
	for _, task := range tasks {
		if !task.IsDone() && pathMatches(task.GetPath(), currentPath) {
			matching = append(matching, task)
		}
	}

	if len(matching) > 0 {
		items := make([]ui.ListItem, len(matching))
		for i, task := range matching {
			sub := ""
			if task.GetEpicID() != nil {
				sub = "part of an epic"
			}
			items[i] = ui.ListItem{Name: task.GetName(), Sub: sub}
		}
		ui.List("TASK", items)
	}

	last, err := config.GetLastPrint()
	if err != nil {
		ui.Error(err.Error())
		return
	}
	if time.Since(last) < 15*time.Minute {
		return
	}

	jobs, err := storage.GetJobs()
	if err != nil {
		ui.Error(err.Error())
		return
	}

	active := make([]*model.Job, 0)
	for _, job := range jobs {
		if !job.IsDone() {
			active = append(active, job)
		}
	}

	if len(active) > 0 {
		items := make([]ui.ListItem, len(active))
		for i, job := range active {
			items[i] = ui.ListItem{Name: job.GetName()}
		}
		ui.List("JOB", items)

		if err := config.SetLastPrint(); err != nil {
			ui.Error(err.Error())
		}
	}
}

func pathJoin(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "/"
		}
		result += p
	}
	return result
}

func pathMatches(taskPath []string, currentPath []string) bool {
	if len(taskPath) != len(currentPath) {
		return false
	}
	for i := range taskPath {
		if taskPath[i] != currentPath[i] {
			return false
		}
	}
	return true
}
