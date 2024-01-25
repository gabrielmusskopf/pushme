package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func openBrowser(url string) {
	cmd := exec.Command("xdg-open", url)
	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not open browser:", err)
	}
}

func isGitRepository() bool {
	_, err := os.Stat(".git")
	return !os.IsNotExist(err)
}

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "This will run a push command from the current branch to remote repository. After the push is completed,\n")
		fmt.Fprint(os.Stderr, "it will ask you if you want to open a merge request for this branch\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	allYes := flag.Bool("y", false, "accept all")
	flag.Parse()

	if !isGitRepository() {
		fmt.Fprintln(os.Stderr, "ERROR: This is not a git repository")
		return
	}

	currentBranchOutput, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not get current branch", err)
		return
	}

	var buff bytes.Buffer
	branch := string(bytes.TrimSpace(currentBranchOutput))
	pushCmd := exec.Command("git", "push", "-u", "origin", branch, "--progress")
	pushCmd.Stderr = io.MultiWriter(&buff, os.Stdout)

	if err := pushCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not run command: %v", err)
		return
	}
	output := strings.TrimSpace(buff.String())
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Everything up-to-date") {
			break
		}
		if strings.Contains(line, "To create a merge request for") || strings.Contains(line, "View merge request for") {
			if scanner.Scan() {
				url := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "remote:"))
				var response string
				if *allYes {
					response = "s"
				} else {
					fmt.Printf("\nDeseja abrir o Merge Request para %s? (s/n): ", branch)
					fmt.Scanln(&response)
				}
				if strings.ToLower(response) == "s" {
					openBrowser(url)
				}
				break
			}
		}
	}

	//TODO: Perguntar se deseja usar target branch

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: ", err)
	}
}
