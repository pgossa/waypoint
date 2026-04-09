package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ── Model ─────────────────────────────────────────────────────────────────────

type Status string

const (
	StatusPending Status = "pending"
	StatusDone    Status = "done"
)

type Task struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Folder    string     `json:"folder"`
	Status    Status     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

// Epic tracks a named group of regular tasks spread across multiple folders.
// Progress is computed by inspecting the actual tasks in TrackedFolders.
// Each tracked folder has an associated task ID so we can find it fast.
type TrackedFolder struct {
	Folder string `json:"folder"`
	TaskID string `json:"task_id"` // ID of the regular Task in this folder
}

type Epic struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Folder         string          `json:"folder"` // parent folder — where epic summary is shown
	TrackedFolders []TrackedFolder `json:"tracked_folders"`
	Status         Status          `json:"status"`  // explicit override: "done" means force-completed
	CreatedAt      time.Time       `json:"created_at"`
	DoneAt         *time.Time      `json:"done_at,omitempty"`
}

// Progress computes done/total by looking up tasks in the store.
func (e *Epic) Progress(tasks []Task) (done, total int) {
	taskByID := make(map[string]*Task, len(tasks))
	for i := range tasks {
		taskByID[tasks[i].ID] = &tasks[i]
	}
	for _, tf := range e.TrackedFolders {
		t, ok := taskByID[tf.TaskID]
		if !ok {
			continue
		}
		total++
		if t.Status == StatusDone {
			done++
		}
	}
	return
}

func (e *Epic) IsComplete(tasks []Task) bool {
	if e.Status == StatusDone {
		return true
	}
	done, total := e.Progress(tasks)
	return total > 0 && done == total
}

// ── Storage ───────────────────────────────────────────────────────────────────

type Store struct {
	Tasks []Task `json:"tasks"`
	Epics []Epic `json:"epics"`
}

func storagePath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "wpt", "tasks.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fatal("cannot determine home directory: %v", err)
	}
	return filepath.Join(home, ".local", "share", "wpt", "tasks.json")
}

func load() Store {
	path := storagePath()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Store{}
	}
	if err != nil {
		fatal("cannot read storage file: %v", err)
	}
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		fatal("storage file is corrupted: %v", err)
	}
	return store
}

func save(store Store) {
	path := storagePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fatal("cannot create storage directory: %v", err)
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		fatal("cannot serialize tasks: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fatal("cannot write storage file: %v", err)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "wpt: "+format+"\n", args...)
	os.Exit(1)
}

func absPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		p = filepath.Join(home, p[2:])
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		fatal("invalid path %q: %v", p, err)
	}
	return abs
}

func pwd() string {
	dir, err := os.Getwd()
	if err != nil {
		fatal("cannot determine current directory: %v", err)
	}
	return dir
}

func generateID() string {
	// Use nanoseconds + a counter suffix to avoid collisions in tight loops.
	return fmt.Sprintf("%x-%d", time.Now().UnixNano(), idCounter())
}

var _idCounter int

func idCounter() int {
	_idCounter++
	return _idCounter
}

func isUnder(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

func looksLikePath(s string) bool {
	return strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../") ||
		strings.HasPrefix(s, "~/") ||
		s == "~"
}

func assertDir(path string) {
	info, err := os.Stat(path)
	if err != nil {
		fatal("path does not exist: %s", path)
	}
	if !info.IsDir() {
		fatal("path is not a directory: %s", path)
	}
}

// parseFlags extracts known flags (long + short alias) from args.
func parseFlags(args []string, aliases map[string]string) (map[string]bool, []string) {
	set := map[string]bool{}
	var rest []string
	for _, a := range args {
		matched := false
		for long, short := range aliases {
			if a == long || (short != "" && a == short) {
				set[long] = true
				matched = true
				break
			}
		}
		if !matched {
			rest = append(rest, a)
		}
	}
	return set, rest
}

// resolvePathAndRest consumes first arg as path if it looks like one, else pwd().
func resolvePathAndRest(args []string) (string, []string) {
	if len(args) > 0 && looksLikePath(args[0]) {
		return absPath(args[0]), args[1:]
	}
	return pwd(), args
}

// taskByID builds an id→*Task index.
func taskByID(tasks []Task) map[string]*Task {
	m := make(map[string]*Task, len(tasks))
	for i := range tasks {
		m[tasks[i].ID] = &tasks[i]
	}
	return m
}

// epicForFolder returns all epics that track the given folder (as a subfolder).
func epicsForFolder(epics []Epic, folder string) []*Epic {
	var out []*Epic
	for i := range epics {
		for _, tf := range epics[i].TrackedFolders {
			if tf.Folder == folder {
				out = append(out, &epics[i])
				break
			}
		}
	}
	return out
}

// taskIDForFolder returns the task ID tracked by an epic for a given folder.
func taskIDForFolder(e *Epic, folder string) string {
	for _, tf := range e.TrackedFolders {
		if tf.Folder == folder {
			return tf.TaskID
		}
	}
	return ""
}

// ── Colors ────────────────────────────────────────────────────────────────────

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorOrange = "\033[38;5;208m"
)

func color(c, s string) string {
	if !isTerminal() {
		return s
	}
	return c + s + colorReset
}

func bold(s string) string   { return color(colorBold, s) }
func gray(s string) string   { return color(colorGray, s) }
func green(s string) string  { return color(colorGreen, s) }
func cyan(s string) string   { return color(colorCyan, s) }
func yellow(s string) string { return color(colorYellow, s) }
func orange(s string) string { return color(colorOrange, s) }

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// progressBar renders: [████░░░░] 3/8 (37%)
func progressBar(done, total int) string {
	if total == 0 {
		return gray("(no tasks)")
	}
	const width = 10
	filled := (done * width) / total
	pct := (done * 100) / total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	barColor := colorYellow
	if pct == 100 {
		barColor = colorGreen
	} else if pct >= 50 {
		barColor = colorCyan
	}
	return fmt.Sprintf("%s %s",
		color(barColor, "["+bar+"]"),
		gray(fmt.Sprintf("%d/%d (%d%%)", done, total, pct)),
	)
}

// ── cmd: add ──────────────────────────────────────────────────────────────────

func cmdAdd(args []string) {
	flags, rest := parseFlags(args, map[string]string{"--recursive": "-r"})
	recursive := flags["--recursive"]

	targetPath, rest := resolvePathAndRest(rest)
	if len(rest) == 0 {
		fatal("usage: wpt add [path] \"task name\" [-r|--recursive]")
	}
	taskName := rest[0]

	assertDir(targetPath)
	store := load()

	if recursive {
		entries, err := os.ReadDir(targetPath)
		if err != nil {
			fatal("cannot read directory: %v", err)
		}
		added := 0
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			subPath := filepath.Join(targetPath, e.Name())
			store.Tasks = append(store.Tasks, Task{
				ID:        generateID(),
				Name:      taskName,
				Folder:    subPath,
				Status:    StatusPending,
				CreatedAt: time.Now(),
			})
			fmt.Printf("  %s %s %s\n", green("✚"), bold(taskName), gray("("+subPath+")"))
			added++
		}
		if added == 0 {
			fmt.Println(yellow("⚠ no subdirectories found"))
		} else {
			fmt.Printf("%s added %d task(s)\n", green("✔"), added)
		}
	} else {
		store.Tasks = append(store.Tasks, Task{
			ID:        generateID(),
			Name:      taskName,
			Folder:    targetPath,
			Status:    StatusPending,
			CreatedAt: time.Now(),
		})
		fmt.Printf("  %s %s %s\n", green("✚"), bold(taskName), gray("("+targetPath+")"))
	}

	save(store)
}

// ── cmd: done ─────────────────────────────────────────────────────────────────

func cmdDone(args []string) {
	targetPath, rest := resolvePathAndRest(args)
	var filter string
	if len(rest) > 0 {
		filter = rest[0]
	}

	store := load()

	var pending []int
	for i, t := range store.Tasks {
		if t.Folder == targetPath && t.Status == StatusPending {
			pending = append(pending, i)
		}
	}
	if len(pending) == 0 {
		fatal("no pending tasks found in %s", targetPath)
	}

	var toMark []int
	if filter == "" {
		if len(pending) > 1 {
			fmt.Fprintf(os.Stderr, "wpt: multiple pending tasks in %s — use an index or name:\n", targetPath)
			for seq, idx := range pending {
				fmt.Fprintf(os.Stderr, "  %s  %s\n",
					cyan(fmt.Sprintf("[%d]", seq+1)),
					store.Tasks[idx].Name,
				)
			}
			os.Exit(1)
		}
		toMark = pending
	} else {
		if n, err := strconv.Atoi(filter); err == nil {
			if n < 1 || n > len(pending) {
				fatal("index %d out of range (1–%d)", n, len(pending))
			}
			toMark = []int{pending[n-1]}
		} else {
			lower := strings.ToLower(filter)
			for _, idx := range pending {
				if strings.Contains(strings.ToLower(store.Tasks[idx].Name), lower) {
					toMark = append(toMark, idx)
				}
			}
			if len(toMark) == 0 {
				fatal("no pending task matches %q in %s", filter, targetPath)
			}
			if len(toMark) > 1 {
				fmt.Fprintf(os.Stderr, "wpt: ambiguous match for %q — be more specific:\n", filter)
				for _, idx := range toMark {
					fmt.Fprintf(os.Stderr, "  • %s\n", store.Tasks[idx].Name)
				}
				os.Exit(1)
			}
		}
	}

	now := time.Now()
	for _, idx := range toMark {
		store.Tasks[idx].Status = StatusDone
		store.Tasks[idx].DoneAt = &now
		fmt.Printf("  %s %s %s\n", green("✔"), bold(store.Tasks[idx].Name), gray("marked done"))
	}

	// Show updated epic progress if this task belongs to any epic.
	idx := taskByID(store.Tasks)
	for i := range store.Epics {
		e := &store.Epics[i]
		for _, marked := range toMark {
			markedID := store.Tasks[marked].ID
			for _, tf := range e.TrackedFolders {
				if tf.TaskID == markedID {
					done, total := e.Progress(store.Tasks)
					fmt.Printf("  %s %s  %s\n", orange("◆"), bold(e.Name), progressBar(done, total))
					if e.IsComplete(store.Tasks) {
						fmt.Printf("  %s epic complete!\n", green("🎉"))
					}
					_ = idx
					break
				}
			}
		}
	}

	save(store)
}

// ── cmd: list ─────────────────────────────────────────────────────────────────

func cmdList(args []string) {
	flags, rest := parseFlags(args, map[string]string{
		"--all":  "-a",
		"--done": "-d",
	})
	showAll := flags["--all"]
	showDone := flags["--done"]

	var prefixFilter string
	if len(rest) > 0 {
		prefixFilter = absPath(rest[0])
	}

	wantStatus := StatusPending
	if showDone {
		wantStatus = StatusDone
	}

	store := load()

	if showAll {
		// ── Global view ───────────────────────────────────────────────────────
		type folderEntry struct {
			tasks []*Task
			epics []*Epic
		}
		folderMap := map[string]*folderEntry{}
		var folderOrder []string

		ensureFolder := func(f string) {
			if _, ok := folderMap[f]; !ok {
				folderMap[f] = &folderEntry{}
				folderOrder = append(folderOrder, f)
			}
		}

		for i := range store.Tasks {
			t := &store.Tasks[i]
			if t.Status != wantStatus {
				continue
			}
			if prefixFilter != "" && !isUnder(t.Folder, prefixFilter) {
				continue
			}
			ensureFolder(t.Folder)
			folderMap[t.Folder].tasks = append(folderMap[t.Folder].tasks, t)
		}

		for i := range store.Epics {
			e := &store.Epics[i]
			complete := e.IsComplete(store.Tasks)
			if showDone && !complete {
				continue
			}
			if !showDone && complete {
				continue
			}
			if prefixFilter != "" && !isUnder(e.Folder, prefixFilter) {
				continue
			}
			ensureFolder(e.Folder)
			folderMap[e.Folder].epics = append(folderMap[e.Folder].epics, e)
		}

		if len(folderOrder) == 0 {
			return
		}

		for fi, folder := range folderOrder {
			fmt.Printf("%s\n", cyan(folder))
			entry := folderMap[folder]
			for i, t := range entry.tasks {
				printTask(*t, i+1)
			}
			for i, e := range entry.epics {
				done, total := e.Progress(store.Tasks)
				printEpicSummary(e, i+1, done, total)
			}
			if fi < len(folderOrder)-1 {
				fmt.Println()
			}
		}

	} else {
		// ── Local view ────────────────────────────────────────────────────────
		targetPath := pwd()

		// Regular tasks at this exact folder.
		seq := 0
		for _, t := range store.Tasks {
			if t.Folder == targetPath && t.Status == wantStatus {
				seq++
				printTask(t, seq)
			}
		}

		// Epics whose parent is this folder.
		epicSeq := 0
		for i := range store.Epics {
			e := &store.Epics[i]
			if e.Folder != targetPath {
				continue
			}
			complete := e.IsComplete(store.Tasks)
			if showDone && !complete {
				continue
			}
			if !showDone && complete {
				continue
			}
			epicSeq++
			done, total := e.Progress(store.Tasks)
			printEpicSummary(e, epicSeq, done, total)
		}

		// Epic membership lines: this folder is a tracked subfolder of some epic.
		if !showDone {
			memberEpics := epicsForFolder(store.Epics, targetPath)
			idx := taskByID(store.Tasks)
			for _, e := range memberEpics {
				if e.IsComplete(store.Tasks) {
					continue
				}
				tid := taskIDForFolder(e, targetPath)
				t, ok := idx[tid]
				if !ok {
					continue
				}
				taskMarker := yellow("●")
				taskStatus := "pending"
				if t.Status == StatusDone {
					taskMarker = green("✔")
					taskStatus = "done"
				}
				done, total := e.Progress(store.Tasks)
				fmt.Printf("  %s %s %s %s  %s\n",
					orange("◆"),
					bold(e.Name),
					gray("›"),
					fmt.Sprintf("%s  %s", taskMarker, bold(t.Name)),
					gray(fmt.Sprintf("(%s · epic %d/%d)", taskStatus, done, total)),
				)
			}
		}
	}
}

func printTask(t Task, idx int) {
	marker := yellow("●")
	nameStr := bold(t.Name)
	if t.Status == StatusDone {
		marker = green("✔")
		nameStr = gray(t.Name)
	}
	fmt.Printf("  %s %s %s\n", marker, gray(fmt.Sprintf("[%d]", idx)), nameStr)
}

func printEpicSummary(e *Epic, idx, done, total int) {
	marker := orange("◆")
	if total > 0 && done == total {
		marker = green("◆")
	}
	fmt.Printf("  %s %s %s  %s\n",
		marker,
		gray(fmt.Sprintf("[%d]", idx)),
		bold(e.Name),
		progressBar(done, total),
	)
}

// ── cmd: epic ─────────────────────────────────────────────────────────────────

func cmdEpic(args []string) {
	if len(args) == 0 {
		epicUsage()
		os.Exit(1)
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "add":
		epicAdd(rest)
	case "task":
		epicTask(rest)
	case "done":
		epicDone(rest)
	case "list", "ls":
		epicList(rest)
	default:
		fmt.Fprintf(os.Stderr, "wpt epic: unknown subcommand %q\n\n", sub)
		epicUsage()
		os.Exit(1)
	}
}

// epicAdd: wpt epic add [path] "name" [-r|--recursive]
// With -r: creates a regular task named "<epic> - <subfolder>" in each subfolder
// and registers each as a tracked folder in the epic.
// Without -r: creates an empty epic; use `wpt epic task` to add tracked folders.
func epicAdd(args []string) {
	flags, rest := parseFlags(args, map[string]string{"--recursive": "-r"})
	recursive := flags["--recursive"]

	targetPath, rest := resolvePathAndRest(rest)
	if len(rest) == 0 {
		fatal("usage: wpt epic add [path] \"epic name\" [-r|--recursive]")
	}
	epicName := rest[0]

	assertDir(targetPath)
	store := load()

	epic := Epic{
		ID:        generateID(),
		Name:      epicName,
		Folder:    targetPath,
		CreatedAt: time.Now(),
	}

	if recursive {
		entries, err := os.ReadDir(targetPath)
		if err != nil {
			fatal("cannot read directory: %v", err)
		}
		added := 0
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			subPath := filepath.Join(targetPath, e.Name())
			taskName := epicName + " - " + e.Name()

			// Skip if a task with this name already exists in the subfolder.
			alreadyExists := false
			for _, t := range store.Tasks {
				if t.Folder == subPath && t.Name == taskName {
					fmt.Printf("  %s skipping %s %s\n",
						yellow("⚠"), bold(taskName), gray("(already exists)"))
					alreadyExists = true
					break
				}
			}
			if alreadyExists {
				continue
			}

			task := Task{
				ID:        generateID(),
				Name:      taskName,
				Folder:    subPath,
				Status:    StatusPending,
				CreatedAt: time.Now(),
			}
			store.Tasks = append(store.Tasks, task)
			epic.TrackedFolders = append(epic.TrackedFolders, TrackedFolder{
				Folder: subPath,
				TaskID: task.ID,
			})
			fmt.Printf("  %s %s %s\n", green("✚"), bold(taskName), gray("("+subPath+")"))
			added++
		}
		if added == 0 {
			fmt.Println(yellow("⚠ no subdirectories found — epic created with no tasks"))
		} else {
			fmt.Printf("  %s epic %s — %d task(s) created\n",
				orange("◆"), bold(epicName), added)
		}
	} else {
		fmt.Printf("  %s epic %s %s\n",
			orange("◆"), bold(epicName), gray("("+targetPath+" — add tasks with `wpt epic task`)"))
	}

	store.Epics = append(store.Epics, epic)
	save(store)
}

// epicTask: wpt epic task [path] <epic-index|name> [subfolder-path]
// Adds an existing folder as a tracked entry in the epic, creating the task.
func epicTask(args []string) {
	targetPath, rest := resolvePathAndRest(args)
	if len(rest) < 1 {
		fatal("usage: wpt epic task [path] <epic-index|name> [subfolder-path]")
	}

	store := load()
	epicIdx := findEpic(store.Epics, targetPath, rest[0])
	e := &store.Epics[epicIdx]

	// Determine the subfolder to track.
	var subPath string
	if len(rest) >= 2 {
		subPath = absPath(rest[1])
	} else {
		// Default: use pwd() as the subfolder to add.
		subPath = pwd()
	}
	assertDir(subPath)

	// Make sure this folder isn't already tracked.
	for _, tf := range e.TrackedFolders {
		if tf.Folder == subPath {
			fatal("folder %s is already tracked by epic %q", subPath, e.Name)
		}
	}

	taskName := e.Name + " - " + filepath.Base(subPath)
	task := Task{
		ID:        generateID(),
		Name:      taskName,
		Folder:    subPath,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	store.Tasks = append(store.Tasks, task)
	e.TrackedFolders = append(e.TrackedFolders, TrackedFolder{
		Folder: subPath,
		TaskID: task.ID,
	})

	done, total := e.Progress(store.Tasks)
	fmt.Printf("  %s %s → epic %s  %s\n",
		green("✚"), bold(taskName), bold(e.Name), progressBar(done, total))

	save(store)
}

// epicDone: wpt epic done [path] <epic-index|name> [subfolder-index|name] [--force|-f]
// --force marks all pending tasks in the epic as done.
// With a subfolder index/name, marks that specific tracked folder's task done.
// Without either, marks the task for the current pwd if it is a tracked folder.
func epicDone(args []string) {
	flags, rest := parseFlags(args, map[string]string{"--force": "-f"})
	force := flags["--force"]

	targetPath, rest := resolvePathAndRest(rest)
	if len(rest) < 1 {
		fatal("usage: wpt epic done [path] <epic-index|name> [subfolder-index|name] [--force|-f]")
	}
	epicFilter := rest[0]
	// Optional second arg: tracked-folder index (1-based) or name fragment.
	var subFilter string
	if len(rest) >= 2 {
		subFilter = rest[1]
	}

	store := load()
	epicIdx := findEpic(store.Epics, targetPath, epicFilter)
	e := &store.Epics[epicIdx]
	idx := taskByID(store.Tasks)

	now := time.Now()

	if force {
		// Mark all pending tracked tasks done, then mark the epic itself done.
		marked := 0
		for _, tf := range e.TrackedFolders {
			t, ok := idx[tf.TaskID]
			if !ok || t.Status == StatusDone {
				continue
			}
			t.Status = StatusDone
			t.DoneAt = &now
			fmt.Printf("  %s %s %s\n", green("✔"), bold(t.Name), gray("marked done"))
			marked++
		}
		e.Status = StatusDone
		e.DoneAt = &now
		done, total := e.Progress(store.Tasks)
		fmt.Printf("  %s %s  %s\n", orange("◆"), bold(e.Name), progressBar(done, total))
		fmt.Printf("  %s epic force-completed!\n", green("🎉"))
		_ = marked
	} else {
		// Resolve which tracked folder to mark done.
		var targetTF *TrackedFolder

		if subFilter != "" {
			// Numeric index into tracked folders (1-based, shown by epic list).
			if n, err := strconv.Atoi(subFilter); err == nil {
				if n < 1 || n > len(e.TrackedFolders) {
					fatal("subfolder index %d out of range (1–%d)", n, len(e.TrackedFolders))
				}
				tf := e.TrackedFolders[n-1]
				targetTF = &tf
			} else {
				// Name fragment match against the base folder name.
				lower := strings.ToLower(subFilter)
				var matches []int
				for i, tf := range e.TrackedFolders {
					if strings.Contains(strings.ToLower(filepath.Base(tf.Folder)), lower) {
						matches = append(matches, i)
					}
				}
				if len(matches) == 0 {
					fatal("no tracked folder matches %q in epic %q", subFilter, e.Name)
				}
				if len(matches) > 1 {
					fmt.Fprintf(os.Stderr, "wpt: ambiguous match for %q — be more specific:\n", subFilter)
					for _, i := range matches {
						fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, e.TrackedFolders[i].Folder)
					}
					os.Exit(1)
				}
				tf := e.TrackedFolders[matches[0]]
				targetTF = &tf
			}
		} else {
			// Default: use pwd as the tracked subfolder.
			currentDir := pwd()

			// If pwd is the epic's own parent folder and the epic has no tracked
			// folders (or none match), mark the epic itself done directly.
			if currentDir == targetPath {
				if len(e.TrackedFolders) == 0 {
					// No tasks — just mark the epic done.
					e.Status = StatusDone
					e.DoneAt = &now
					fmt.Printf("  %s %s %s\n", green("◆"), bold(e.Name), gray("marked done"))
					fmt.Printf("  %s epic complete!\n", green("🎉"))
					save(store)
					return
				}
				// Has tracked folders — show them so the user can pick.
				fmt.Fprintf(os.Stderr, "wpt: use a subfolder index/name or --force to complete all:\n")
				for j, tf := range e.TrackedFolders {
					t, ok := idx[tf.TaskID]
					marker := yellow("·")
					if ok && t.Status == StatusDone {
						marker = green("✔")
					}
					fmt.Fprintf(os.Stderr, "  %s %s  %s\n",
						marker, cyan(fmt.Sprintf("[%d]", j+1)), filepath.Base(tf.Folder))
				}
				os.Exit(1)
			}

			// pwd is a subfolder — find it in tracked folders.
			for i, tf := range e.TrackedFolders {
				if tf.Folder == currentDir {
					tmp := e.TrackedFolders[i]
					targetTF = &tmp
					break
				}
			}
			if targetTF == nil {
				fmt.Fprintf(os.Stderr, "wpt: current directory is not tracked by epic %q\n", e.Name)
				fmt.Fprintf(os.Stderr, "     use an index or name, or --force to complete all:\n")
				for j, tf := range e.TrackedFolders {
					t, ok := idx[tf.TaskID]
					marker := yellow("·")
					if ok && t.Status == StatusDone {
						marker = green("✔")
					}
					fmt.Fprintf(os.Stderr, "  %s %s  %s\n",
						marker, cyan(fmt.Sprintf("[%d]", j+1)), filepath.Base(tf.Folder))
				}
				os.Exit(1)
			}
		}

		if targetTF == nil {
			// Should not happen but guard just in case.
			fatal("could not resolve a tracked folder to mark done")
		}

		t, ok := idx[targetTF.TaskID]
		if !ok {
			fatal("tracked task not found (storage may be corrupted)")
		}
		if t.Status == StatusDone {
			fatal("task %q is already done", t.Name)
		}
		t.Status = StatusDone
		t.DoneAt = &now
		fmt.Printf("  %s %s %s\n", green("✔"), bold(t.Name), gray("marked done"))

		done, total := e.Progress(store.Tasks)
		fmt.Printf("  %s %s  %s\n", orange("◆"), bold(e.Name), progressBar(done, total))
		if e.IsComplete(store.Tasks) {
			e.Status = StatusDone
			e.DoneAt = &now
			fmt.Printf("  %s epic complete!\n", green("🎉"))
		}
	}

	save(store)
}

// epicList: wpt epic list [-a|--all] [-d|--done] [path]
// Shows epics with their tracked folders and task statuses.
func epicList(args []string) {
	flags, rest := parseFlags(args, map[string]string{
		"--all":  "-a",
		"--done": "-d",
	})
	showAll := flags["--all"]
	showDone := flags["--done"]

	var prefixFilter string
	if len(rest) > 0 {
		prefixFilter = absPath(rest[0])
	}

	targetPath := pwd()
	store := load()
	idx := taskByID(store.Tasks)

	lastFolder := ""
	for i := range store.Epics {
		e := &store.Epics[i]

		if showAll {
			if prefixFilter != "" && !isUnder(e.Folder, prefixFilter) {
				continue
			}
		} else {
			if e.Folder != targetPath {
				continue
			}
		}

		complete := e.IsComplete(store.Tasks)
		if showDone && !complete {
			continue
		}
		if !showDone && complete {
			continue
		}

		if showAll && e.Folder != lastFolder {
			if lastFolder != "" {
				fmt.Println()
			}
			fmt.Printf("%s\n", cyan(e.Folder))
			lastFolder = e.Folder
		}

		done, total := e.Progress(store.Tasks)
		marker := orange("◆")
		if complete {
			marker = green("◆")
		}
		fmt.Printf("  %s %s  %s\n", marker, bold(e.Name), progressBar(done, total))

		for j, tf := range e.TrackedFolders {
			t, ok := idx[tf.TaskID]
			subMarker := yellow("·")
			subName := bold(filepath.Base(tf.Folder))
			if ok && t.Status == StatusDone {
				subMarker = green("✔")
				subName = gray(filepath.Base(tf.Folder))
			}
			fmt.Printf("      %s %s %s %s\n",
				subMarker,
				gray(fmt.Sprintf("[%d]", j+1)),
				subName,
				gray("("+tf.Folder+")"),
			)
		}
	}
}

// findEpic locates an epic by its parent folder + name/index filter.
func findEpic(epics []Epic, folder, filter string) int {
	var candidates []int
	for i, e := range epics {
		if e.Folder == folder {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		fatal("no epics found in %s", folder)
	}

	if n, err := strconv.Atoi(filter); err == nil {
		if n < 1 || n > len(candidates) {
			fatal("epic index %d out of range (1–%d)", n, len(candidates))
		}
		return candidates[n-1]
	}

	lower := strings.ToLower(filter)
	var matches []int
	for _, idx := range candidates {
		if strings.Contains(strings.ToLower(epics[idx].Name), lower) {
			matches = append(matches, idx)
		}
	}
	if len(matches) == 0 {
		fatal("no epic matches %q in %s", filter, folder)
	}
	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "wpt: ambiguous epic match for %q — be more specific:\n", filter)
		for _, idx := range matches {
			fmt.Fprintf(os.Stderr, "  • %s\n", epics[idx].Name)
		}
		os.Exit(1)
	}
	return matches[0]
}

// ── Usage ─────────────────────────────────────────────────────────────────────

func usage() {
	fmt.Print(`wpt — waypoint task tracker

Commands:
  wpt add [path] "name" [-r|--recursive]         Add a task (-r: one per subfolder)
  wpt done [path] [index|name]                   Mark a task done
  wpt list [-a|--all] [-d|--done] [path]         List tasks and epics
  wpt epic <subcommand>                          Manage epics

Epic subcommands:
  wpt epic add  [path] "name" [-r]               Create epic (-r: task per subfolder)
  wpt epic task [path] <epic> [subfolder]        Add a folder/task to an epic
  wpt epic done [path] <epic> [--force|-f]       Mark epic task(s) done
  wpt epic list [-a|--all] [-d|--done] [path]    List epics with task detail

Flags:
  -r, --recursive   One task per direct subfolder
  -a, --all         Show items across all directories
  -d, --done        Show completed items instead of pending
  -f, --force       (epic done) mark all pending tasks in the epic done

Examples:
  wpt add "fix auth bug"                          Add task in current dir
  wpt add ~/apps "deploy" -r                      Add task to every subfolder
  wpt done                                        Mark done (fails if ambiguous)
  wpt done 2                                      Mark task #2 done
  wpt list                                        Pending tasks here (chpwd hook)
  wpt list -a                                     All pending tasks everywhere
  wpt list -a ~/my-apps                           All pending under ~/my-apps
  wpt list -d                                     Completed tasks in current dir

  wpt epic add ~/my-apps "Update deps" -r         Epic + task per subfolder
  wpt epic add "Q4 release"                       Empty epic, add folders manually
  wpt epic task "Q4" ~/my-apps/app4               Track app4 in epic "Q4"
  wpt epic done ~/my-apps "Update deps"           Mark current folder's task done
  wpt epic done ~/my-apps "Update deps" --force   Mark ALL tasks in epic done
  wpt epic list                                   Epics here (with task detail)
  wpt epic list -a                                All epics everywhere

Zsh integration:
  chpwd() { wpt list }
`)
}

func epicUsage() {
	fmt.Print(`wpt epic — manage epics (groups of regular tasks across folders)

Subcommands:
  add  [path] "name" [-r]            Create epic (-r: task per subfolder)
  task [path] <epic> [subfolder]     Track a folder in an existing epic
  done [path] <epic> [-f|--force]    Mark task(s) done (-f: force all)
  list [-a|--all] [-d|--done]        List epics with tracked folder detail
`)
}

// ── Entry point ───────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}

	cmd, rest := os.Args[1], os.Args[2:]
	switch cmd {
	case "add":
		cmdAdd(rest)
	case "done":
		cmdDone(rest)
	case "list", "ls":
		cmdList(rest)
	case "epic":
		cmdEpic(rest)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "wpt: unknown command %q\n\n", cmd)
		usage()
		os.Exit(1)
	}
}
