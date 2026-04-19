package task

import (
	"fmt"
	"os"

	"waypoint/cmd/epic"
	"waypoint/storage"
	"waypoint/ui"
)

func taskBound(args []string) {
	if len(args) != 2 {
		taskBoundUsage()
		return
	}

	// find task
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
	if task.GetEpicID() != nil {
		ui.Error("task is already bound to an epic: " + task.GetName())
		os.Exit(1)
	}

	// find epic
	epics, err := epic.GetActiveEpics()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	epicIdx, epicMatches := epic.FindEpic(args[1], epics)
	if len(epicMatches) == 0 {
		ui.Error("no active epic found matching: " + args[1])
		os.Exit(1)
	}
	if len(epicMatches) > 1 {
		ui.Error("multiple epics match, be more specific:")
		for _, i := range epicMatches {
			epic.PrintEpicWithIndex(i, epics[i])
		}
		os.Exit(1)
	}

	epic := epics[epicIdx]

	// bind
	task.SetEpicID(epic.GetID())
	if err := epic.AddTaskID(epic.GetID()); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := storage.SaveTask(task); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	if err := storage.SaveEpic(epic); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("task '%s' bound to epic '%s'", task.GetName(), epic.GetName()))
}

func taskBoundUsage() {
	ui.Info("Usage: wpt task bound <task> <epic>")
	ui.Info("  <task>  Full or partial task name")
	ui.Info("  <epic>  Full or partial epic name")
}
