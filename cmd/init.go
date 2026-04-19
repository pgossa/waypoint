package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"waypoint/config"
	"waypoint/ui"

	"github.com/charmbracelet/huh"
)

func Init(args []string) {
	configureStorage()
	configureShell()
	ui.Success("waypoint is ready to use!")
}

func configureStorage() {
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
	if err := config.Save(&config.Config{Storage: config.StorageType(storage)}); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	ui.Success("storage set to: " + storage)
}

func configureShell() {
	var setup bool
	form := huh.NewConfirm().
		Title("Would you like to set up shell integration? (adds wptg alias and wpt list on cd)").
		Value(&setup)
	if err := form.Run(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	if !setup {
		return
	}

	shell := os.Getenv("SHELL")
	var rcFile string
	var snippet string

	switch {
	case strings.Contains(shell, "zsh"):
		rcFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
		snippet = "\n# waypoint\nwptg() { cd \"$(wpt goto $@)\" }\nchpwd() { wpt cd }\n"
	case strings.Contains(shell, "bash"):
		rcFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
		snippet = "\n# waypoint\nwptg() { cd \"$(wpt goto $@)\" }\nPROMPT_COMMAND=\"wpt cd\"\n"
	default:
		ui.Error("unsupported shell: " + shell)
		ui.Info("please add manually: wptg() { cd \"$(wpt goto $@)\" }")
		return
	}

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString(snippet); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("shell integration added to " + rcFile)
	ui.Info("restart your shell or run: source " + rcFile)
}
