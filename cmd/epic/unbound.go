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

func epicUnbound(args []string) {
	if len(args) < 1 || len(args) > 2 {
		epicUnboundUsage()
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

	if len(epic.GetTasksID()) == 0 {
		ui.Error("epic has no bound tasks: " + epic.GetName())
		os.Exit(1)
	}

	// fetch tasks
	allTasks, err := storage.GetTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	taskMap := make(map[uuid.UUID]*model.Task)
	for _, t := range allTasks {
		taskMap[t.GetID()] = t
	}

	var choice uuid.UUID

	if len(args) == 2 {
		// find task by name from second argument
		input := strings.ToLower(args[1])
		var found []uuid.UUID
		for _, tid := range epic.GetTasksID() {
			t, ok := taskMap[tid]
			if !ok {
				continue
			}
			if strings.Contains(strings.ToLower(t.GetName()), input) {
				found = append(found, tid)
			}
		}
		if len(found) == 0 {
			ui.Error("no bound task found matching: " + args[1])
			os.Exit(1)
		}
		if len(found) > 1 {
			ui.Error("multiple tasks match, be more specific:")
			for _, tid := range found {
				t := taskMap[tid]
				ui.Info(fmt.Sprintf("  [ ] %-30s %s", t.GetName(), strings.Join(t.GetPath(), "/")))
			}
			os.Exit(1)
		}
		choice = found[0]
	} else {
		// interactive selector
		options := make([]huh.Option[uuid.UUID], 0)
		for _, tid := range epic.GetTasksID() {
			t, ok := taskMap[tid]
			if !ok {
				continue
			}
			status := " "
			if t.IsDone() {
				status = "✓"
			}
			label := fmt.Sprintf("[%s] %-30s %s", status, t.GetName(), strings.Join(t.GetPath(), "/"))
			options = append(options, huh.NewOption(label, tid))
		}

		form := huh.NewSelect[uuid.UUID]().
			Title(fmt.Sprintf("Which task to unbound from '%s'?", epic.GetName())).
			Options(options...).
			Value(&choice)
		if err := form.Run(); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
	}

	// confirm
	var confirm bool
	confirmForm := huh.NewConfirm().
		Title(fmt.Sprintf("Unbound task '%s' from epic '%s'?", taskMap[choice].GetName(), epic.GetName())).
		Value(&confirm)
	if err := confirmForm.Run(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	if !confirm {
		ui.Info("aborted.")
		return
	}

	epic.RemoveTaskID(choice)
	if err := storage.SaveEpic(epic); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	task := taskMap[choice]
	task.UnsetEpicID()
	if err := storage.SaveTask(task); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("task '%s' unbound from epic '%s'", task.GetName(), epic.GetName()))
}

func epicUnboundUsage() {
	ui.Info("Usage: wpt epic unbound <epic> [task]")
	ui.Info("  <epic>  Full or partial epic name")
	ui.Info("  [task]  Optional full or partial task name, interactive selector if omitted")
}
