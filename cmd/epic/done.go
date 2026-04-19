package epic

import (
	"fmt"
	"os"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"

	"github.com/charmbracelet/huh"
	"github.com/google/uuid"
)

func epicDone(args []string) {
	if len(args) != 1 {
		epicDoneUsage()
		return
	}

	epics, err := GetActiveEpics()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	idx, matches := FindEpic(args[0], epics)

	if len(matches) == 0 {
		ui.Error("no active epic found matching: " + args[0])
		os.Exit(1)
	}
	if len(matches) > 1 {
		ui.Error("multiple epics match, be more specific:")
		for _, i := range matches {
			PrintEpicWithIndex(i, epics[i])
		}
		os.Exit(1)
	}

	epic := epics[idx]

	// fetch subtasks
	allTasks, err := storage.GetTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	taskMap := make(map[uuid.UUID]*model.Task)
	for _, t := range allTasks {
		taskMap[t.GetID()] = t
	}

	pending := make([]*model.Task, 0)
	for _, tid := range epic.GetTasksID() {
		if t, ok := taskMap[tid]; ok && !t.IsDone() {
			pending = append(pending, t)
		}
	}

	if len(pending) > 0 {
		// show pending subtasks
		ui.Info(fmt.Sprintf("%d subtask(s) are still pending:", len(pending)))
		for _, t := range pending {
			ui.Info(fmt.Sprintf("  [ ] %-30s %s", t.GetName(), strings.Join(t.GetPath(), "/")))
		}

		var confirm bool
		form := huh.NewConfirm().
			Title("Mark all pending subtasks and epic as done?").
			Value(&confirm)
		if err := form.Run(); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		if !confirm {
			ui.Info("aborted.")
			return
		}

		// mark all pending subtasks as done
		for _, t := range pending {
			t.MarkDone()
			if err := storage.SaveTask(t); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}
		}
	}

	epic.MarkDone()
	if err := storage.SaveEpic(epic); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("epic done: " + epic.GetName())
}

func epicDoneUsage() {
	ui.Info("Usage: wpt epic done <index|name>")
	ui.Info("  <index>  The epic index from 'wpt epic list'")
	ui.Info("  <name>   Full or partial epic name")
}
