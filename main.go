package main

import (
	"fmt"
	"os"

	"waypoint/cmd"
	"waypoint/cmd/epic"
	"waypoint/cmd/job"
	"waypoint/cmd/task"
	"waypoint/config"
	"waypoint/ui"
)

var version = "dev"

func main() {
	ui.Init()
	if !config.IsInitialized() {
		ui.Info("Welcome to waypoint! Let's get you set up.")
		cmd.Init([]string{})
	}
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}

	command, rest := os.Args[1], os.Args[2:]
	switch command {
	case "job", "j":
		job.Job(rest)
	case "list", "ls", "l":
		cmd.List(rest)
	case "cd":
		cmd.Cd(rest)
	case "task", "t":
		task.Task(rest)
	case "epic", "e":
		epic.Epic(rest)
	case "next", "n":
		cmd.Next(rest)
	case "config", "conf", "c":
		cmd.Configure(rest)
	case "help", "--help", "-h":
		usage()
	case "version", "v", "--version", "-v":
		printVersion()
	default:
		ui.Error(fmt.Sprintf("wpt: unknown command %q\n\n", command))
		usage()
		os.Exit(1)
	}
}

func usage() {
}

func printVersion() {
	ui.Info(fmt.Sprint("waypoint-", version))
}
