package cmd

import (
	"fmt"
	"os"

	"waypoint/config"
	"waypoint/ui"

	"github.com/charmbracelet/huh"
)

func Configure(args []string) {
	if len(args) < 1 {
		configureUsage()
		os.Exit(1)
	}
	command, rest := args[0], args[1:]
	switch command {
	case "storage", "--storage":
		configDB(rest)
	case "help", "--help", "-h":
		configureUsage()
	default:
		ui.Error(fmt.Sprintf("wpt config: unknown command %q", command))
		configureUsage()
		os.Exit(1)
	}
}

func configDB(args []string) {
	if len(args) > 0 {
		// direct set via argument
		switch args[0] {
		case "json":
			saveStorage(config.StorageJSON)
		case "sqlite":
			saveStorage(config.StorageSQLite)
		default:
			ui.Error("unknown storage type: " + args[0])
			configureUsage()
			os.Exit(1)
		}
		return
	}

	// interactive selection
	var storage string
	form := huh.NewSelect[string]().
		Title("Which storage backend would you like to use?").
		Options(
			huh.NewOption("JSON (simple, no dependencies)", "json"),
			huh.NewOption("SQLite (faster, better for large data)", "sqlite"),
		).
		Value(&storage)
	if err := form.Run(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	saveStorage(config.StorageType(storage))
}

func saveStorage(s config.StorageType) {
	cfg, err := config.Load()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	cfg.Storage = s
	if err := config.Save(cfg); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	ui.Success("storage set to: " + string(s))
}

func configureUsage() {
	ui.Info("Usage: wpt config <command>")
	ui.Info("")
	ui.Info("Commands:")
	ui.Info("  storage [json|sqlite]  Set the storage backend")
}
