package task

import (
	"os"

	"waypoint/cmd/epic"
	"waypoint/storage"
	"waypoint/ui"
)

func taskUnbound(args []string) {
	if len(args) != 1 {
		taskUnboundUsage()
		return
	}

	tasks, err := getActiveTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	taskIdx, taskMatches := findTask(args[0], tasks)
	if len(taskMatches) == 0 {
		ui.Error("no active task found matching: " + args[0])
		os.Exit(1)
	}
	if len(taskMatches) > 1 {
		ui.Error("multiple tasks match, be more specific:")
		for _, i := range taskMatches {
			printTaskWithIndex(i, tasks[i])
		}
		os.Exit(1)
	}

	task := tasks[taskIdx]
	if task.GetEpicID() == nil {
		ui.Error("task is not bound to any epic: " + task.GetName())
		os.Exit(1)
	}

	epicID := *task.GetEpicID()
	task.UnsetEpicID()

	// remove task from epic too
	epics, err := epic.GetActiveEpics()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	for _, epic := range epics {
		if epic.GetID() == epicID {
			epic.RemoveTaskID(task.GetID())
			if err := storage.SaveEpic(epic); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}
			break
		}
	}

	if err := storage.SaveTask(task); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("task unbound: " + task.GetName())
}

func taskUnboundUsage() {
	ui.Info("Usage: wpt task unbound <name>")
	ui.Info("  <name>  Full or partial task name")
}
