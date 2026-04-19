package task

import (
	"os"

	"waypoint/storage"
	"waypoint/ui"
)

func taskDone(args []string) {
	if len(args) != 1 {
		taskDoneUsage()
		return
	}

	tasks, err := getActiveTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	idx, matches := findTask(args[0], tasks)

	if idx == -1 && len(matches) == 0 {
		ui.Error("no task found matching: " + args[0])
		os.Exit(1)
	}

	if len(matches) > 1 {
		ui.Error("multiple tasks match, use the index instead:")
		for _, i := range matches {
			printTaskWithIndex(i, tasks[i])
		}
		os.Exit(1)
	}

	task := tasks[idx]
	if task.IsDone() {
		ui.Error("task is already done: " + task.GetName())
		os.Exit(1)
	}

	task.MarkDone()
	if err := storage.SaveTask(task); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("task done: " + task.GetName())
}

func taskDoneUsage() {
	ui.Info("Usage: wpt task done <index|name>")
	ui.Info("  <index>  The task index from 'wpt task list'")
	ui.Info("  <name>   Full or partial task name")
}
