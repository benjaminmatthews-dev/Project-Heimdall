package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"heimdall/runner"
)

const defaultRunnersDir = "/app/conf/runners"

func runnersDir() string {
	if d := os.Getenv("HEIMDALL_RUNNERS_DIR"); d != "" {
		return d
	}
	return defaultRunnersDir
}

func main() {
	dir := runnersDir()
	groups, err := runner.LoadFromDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading runners from %s: %v\n", dir, err)
		os.Exit(1)
	}
	if len(groups) == 0 {
		fmt.Fprintf(os.Stderr, "No runners found in %s\n", dir)
		os.Exit(1)
	}

	// Step 1: select group.
	sortedGroups := runner.SortedGroups(groups)
	fmt.Println("\n=== Heimdall ===")
	fmt.Println("Select a group:")
	for i, g := range sortedGroups {
		fmt.Printf("  [%d] %s  (%d scripts)\n", i+1, g, len(groups[g]))
	}
	reader := bufio.NewReader(os.Stdin)
	groupChoice, err := promptInt(reader, "Group", 1, len(sortedGroups))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading group selection: %v\n", err)
		os.Exit(1)
	}
	groupIdx := groupChoice - 1
	selectedGroup := sortedGroups[groupIdx]
	runners := groups[selectedGroup]

	// Step 2: select script.
	fmt.Printf("\nGroup: %s\n", selectedGroup)
	for i, r := range runners {
		fmt.Printf("  [%d] %s\n", i+1, r.Name)
		if r.Description != "" {
			fmt.Printf("      %s\n", r.Description)
		}
	}
	runnerChoice, err := promptInt(reader, "Script", 1, len(runners))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading script selection: %v\n", err)
		os.Exit(1)
	}
	runnerIdx := runnerChoice - 1
	selected := runners[runnerIdx]

	// Step 3: confirm.
	fmt.Printf("\nReady to run: %s\n", selected.Name)
	fmt.Printf("  %s\n", selected.ScriptPath)
	ok, err := promptConfirm(reader, "Proceed?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading confirmation: %v\n", err)
		os.Exit(1)
	}
	if !ok {
		fmt.Println("Aborted.")
		os.Exit(0)
	}

	// Step 4: execute and stream output.
	if err := execute(selected); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func execute(r runner.Runner) error {
	cmd := exec.Command("sh", "-c", r.ScriptPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func promptInt(reader *bufio.Reader, label string, min, max int) (int, error) {
	for {
		fmt.Printf("%s [%d-%d]: ", label, min, max)
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		n, convErr := strconv.Atoi(line)
		if convErr == nil && n >= min && n <= max {
			return n, nil
		}
		if readErr == io.EOF {
			return 0, io.EOF
		}
		fmt.Printf("  Please enter a number between %d and %d.\n", min, max)
	}
}

func promptConfirm(reader *bufio.Reader, label string) (bool, error) {
	for {
		fmt.Printf("%s [y/N]: ", label)
		line, err := reader.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "y", "yes":
			return true, nil
		case "n", "no", "":
			return false, nil
		}
		if err == io.EOF {
			return false, io.EOF
		}
	}
}
