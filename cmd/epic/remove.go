package epic

import (
	"fmt"
	"os"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"

	"github.com/charmbracelet/huh"
)

func epicRemove(args []string) {
	if len(args) != 1 {
		epicRemoveUsage()
		return
	}

	epics, err := GetSortedEpics()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	_, matches := FindEpic(args[0], epics)

	if len(matches) == 0 {
		ui.Error("no epic found matching: " + args[0])
		os.Exit(1)
	}

	var selected *model.Epic

	if len(matches) == 1 {
		selected = epics[matches[0]]
	} else {
		options := make([]huh.Option[int], len(matches))
		for i, idx := range matches {
			epic := epics[idx]
			status := " "
			if epic.IsDone() {
				status = "✓"
			}
			label := fmt.Sprintf("[%s] %-30s %s", status, epic.GetName(), strings.Join(epic.GetPath(), "/"))
			options[i] = huh.NewOption(label, idx)
		}

		var choice int
		form := huh.NewSelect[int]().
			Title("Multiple epics match, which one do you want to remove?").
			Options(options...).
			Value(&choice)

		if err := form.Run(); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		selected = epics[choice]
	}

	// warn if epic has subtasks
	if len(selected.GetTasksID()) > 0 {
		var confirmTasks bool
		form := huh.NewConfirm().
			Title(fmt.Sprintf("Also remove %d linked subtask(s)?", len(selected.GetTasksID()))).
			Value(&confirmTasks)
		if err := form.Run(); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		if confirmTasks {
			for _, tid := range selected.GetTasksID() {
				if err := storage.RemoveTask(tid); err != nil {
					ui.Error(err.Error())
					os.Exit(1)
				}
			}
		}
	}

	if err := storage.RemoveEpic(selected.GetID()); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("epic removed: " + selected.GetName())
}

func epicRemoveUsage() {
	ui.Info("Usage: wpt epic remove <name>")
	ui.Info("  <name>  Full or partial epic name")
}
