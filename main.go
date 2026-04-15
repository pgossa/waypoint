package main

import (
	"fmt"
	"os"
	"waypoint/cmd"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}

	command, rest := os.Args[1], os.Args[2:]
	switch command {
	case "job", "j":
		cmd.Job(rest)
	case "list", "ls", "l":
		cmd.List(rest)
	case "task", "t":
		cmd.Task(rest)
	case "epic", "e":
		cmd.Epic(rest)
	case "help", "--help", "-h":
		usage()
	case "version", "v", "--version", "-v":
		printVersion()
	default:
		fmt.Fprintf(os.Stderr, "wpt: unknown command %q\n\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {

}

func printVersion() {
	fmt.Println("wpt", version)
}
