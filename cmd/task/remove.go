package task

import (
	"fmt"
	"os"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"

	"github.com/charmbracelet/huh"
)

func taskRemove(args []string) {
	if len(args) != 1 {
		taskRemoveUsage()
		return
	}

	tasks, err := getSortedTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	_, matches := findTask(args[0], tasks)

	if len(matches) == 0 {
		ui.Error("no task found matching: " + args[0])
		os.Exit(1)
	}

	var selected *model.Task

	if len(matches) == 1 {
		selected = tasks[matches[0]]
	} else {
		options := make([]huh.Option[int], len(matches))
		for i, idx := range matches {
			task := tasks[idx]
			status := " "
			if task.IsDone() {
				status = "✓"
			}
			label := fmt.Sprintf("[%s] %-30s %s", status, task.GetName(), strings.Join(task.GetPath(), "/"))
			options[i] = huh.NewOption(label, idx)
		}

		var choice int
		form := huh.NewSelect[int]().
			Title("Multiple tasks match, which one do you want to remove?").
			Options(options...).
			Value(&choice)

		if err := form.Run(); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		selected = tasks[choice]
	}

	var confirm bool
	confirmForm := huh.NewConfirm().
		Title(fmt.Sprintf("Remove task '%s'?", selected.GetName())).
		Value(&confirm)
	if err := confirmForm.Run(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	if !confirm {
		ui.Info("aborted.")
		return
	}

	// if task is bound to an epic, remove it from the epic too
	if selected.GetEpicID() != nil {
		epics, err := storage.GetEpics()
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		for _, epic := range epics {
			if epic.GetID() == *selected.GetEpicID() {
				epic.RemoveTaskID(selected.GetID())
				if err := storage.SaveEpic(epic); err != nil {
					ui.Error(err.Error())
					os.Exit(1)
				}
				break
			}
		}
	}

	if err := storage.RemoveTask(selected.GetID()); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("task removed: " + selected.GetName())
}

func taskRemoveUsage() {
	ui.Info("Usage: wpt task remove <name>")
	ui.Info("  <name>  Full or partial task name")
}
