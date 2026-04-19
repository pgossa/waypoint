package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"waypoint/storage"
	"waypoint/ui"

	"github.com/charmbracelet/huh"
)

func Goto(args []string) {
	tasks, err := storage.GetTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	epics, err := storage.GetEpics()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	// build options from active tasks and epics
	type entry struct {
		name string
		path []string
	}

	entries := make([]entry, 0)

	for _, t := range tasks {
		if !t.IsDone() {
			entries = append(entries, entry{
				name: fmt.Sprintf("[task] %s", t.GetName()),
				path: t.GetPath(),
			})
		}
	}

	for _, e := range epics {
		if !e.IsDone() {
			entries = append(entries, entry{
				name: fmt.Sprintf("[epic] %s", e.GetName()),
				path: e.GetPath(),
			})
		}
	}

	if len(entries) == 0 {
		ui.Error("no active tasks or epics found")
		os.Exit(1)
	}

	// filter by partial name if argument provided
	if len(args) == 1 {
		input := strings.ToLower(args[0])
		filtered := make([]entry, 0)
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.name), input) {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) == 0 {
			ui.Error("no active tasks or epics found matching: " + args[0])
			os.Exit(1)
		}
		entries = filtered
	}

	// sort by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	var selected entry

	if len(entries) == 1 {
		selected = entries[0]
	} else {
		options := make([]huh.Option[int], len(entries))
		for i, e := range entries {
			options[i] = huh.NewOption(
				fmt.Sprintf("%-40s %s", e.name, strings.Join(e.path, "/")),
				i,
			)
		}

		var choice int
		form := huh.NewSelect[int]().
			Title("Where do you want to go?").
			Options(options...).
			Value(&choice)
		if err := form.Run(); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		selected = entries[choice]
	}

	// print path for shell to cd into
	fmt.Println("/" + strings.Join(selected.path, "/"))
}
