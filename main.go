package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

var allYes bool

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

func gitCurrentBranch() (string, error) {
	output, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(output)), nil
}

func gitPreviousBranch() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--symbolic-full-name", "--abbrev-ref=loose", "@{-1}").Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(output)), nil
}

func gitPush(branch string) (string, error) {
	var buff bytes.Buffer
	pushCmd := exec.Command("git", "push", "-u", "origin", branch, "--progress")
	pushCmd.Stderr = io.MultiWriter(&buff, os.Stdout)

	if err := pushCmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buff.String()), nil
}

func getAnswerBool(question string) bool {
	var response string
	if allYes {
		response = "y"
	} else {
		fmt.Print(question)
		fmt.Scanln(&response)
	}

	return response == "y"
}

func appendPreviousBranch(u *string) {
	previousBranch, _ := gitPreviousBranch()
	//TODO: Fix error occuring when does not exist any previous branch. Maybe a better solution
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, "ERROR: Could not get previous branch", err)
	//	return
	//}

	if previousBranch != "" {
		usePreviousAsTarget := getAnswerBool(fmt.Sprintf("Do you want to use last branch (%s) as target? (y/n): ", previousBranch))
		if usePreviousAsTarget {
			*u += "&merge_request" + url.QueryEscape("[target_branch]") + "=" + url.QueryEscape(previousBranch)
		}
	}
}

func main() {
	flag.BoolVar(&allYes, "y", false, "Accept all. This means open a Merge Request and use previous branch as target")
	flag.Parse()

	if !isGitRepository() {
		fmt.Fprintln(os.Stderr, "ERROR: This is not a git repository")
		return
	}


	currentBranch, err := gitCurrentBranch()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not get current branch", err)
		return
	}

	push, err := gitPush(currentBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not push: %v", err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(push))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Everything up-to-date") {
			break
		}
		if strings.Contains(line, "To create a merge request for") || strings.Contains(line, "View merge request for") {
			if scanner.Scan() {
				u := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "remote:"))
				openMerge := getAnswerBool(fmt.Sprintf("\nDo you want to open a new Merge Request to %s? (y/n): ", currentBranch))
				if openMerge {
                    appendPreviousBranch(&u)
					openBrowser(u)
				}
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: ", err)
	}

}
