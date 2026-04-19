package job

import (
	"fmt"
	"os"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"

	"github.com/charmbracelet/huh"
)

func jobRemove(args []string) {
	if len(args) != 1 {
		jobRemoveUsage()
		return
	}

	jobs, err := getSortedJobs()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	_, matches := findJob(args[0], jobs)

	if len(matches) == 0 {
		ui.Error("no job found matching: " + args[0])
		os.Exit(1)
	}

	var selected *model.Job

	if len(matches) == 1 {
		selected = jobs[matches[0]]
	} else {
		options := make([]huh.Option[int], len(matches))
		for i, idx := range matches {
			job := jobs[idx]
			status := " "
			if job.IsDone() {
				status = "✓"
			}
			label := fmt.Sprintf("[%s] %s", status, job.GetName())
			options[i] = huh.NewOption(label, idx)
		}

		var choice int
		form := huh.NewSelect[int]().
			Title("Multiple jobs match, which one do you want to remove?").
			Options(options...).
			Value(&choice)

		if err := form.Run(); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		selected = jobs[choice]
	}

	var confirm bool
	confirmForm := huh.NewConfirm().
		Title(fmt.Sprintf("Remove job '%s'?", selected.GetName())).
		Value(&confirm)
	if err := confirmForm.Run(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	if !confirm {
		ui.Info("aborted.")
		return
	}

	if err := storage.RemoveJob(selected.GetID()); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("job removed: " + selected.GetName())
}

func jobRemoveUsage() {
	ui.Info("Usage: wpt job remove <name>")
	ui.Info("  <name>  Full or partial job name")
}
