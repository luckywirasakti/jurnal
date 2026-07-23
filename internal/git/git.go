package git

import (
	"bytes"
	"os/exec"
	"strings"
)

// Command runs git commands synchronously returning stdout.
func Command(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// IsRepository checks if directory is inside a git working tree.
func IsRepository() bool {
	_, err := Command("rev-parse", "--git-dir")
	return err == nil
}

// GetStatus returns the current working directory raw status.
func GetStatus() string {
	out, _ := Command("status")
	return out
}

// GetDiff gathers both untracked and unstaged local modifications.
func GetDiff() string {
	diff, _ := Command("diff")
	untracked, _ := Command("status", "--porcelain")
	if untracked != "" {
		return untracked + "\n\n" + diff
	}
	return diff
}

// GetCurrentBranch returns the active branch. Defaults to master/main.
func GetCurrentBranch() string {
	branch, err := Command("branch", "--show-current")
	if err != nil || branch == "" {
		return "main"
	}
	return branch
}

// Checkout switches the active branch or creates a new one.
func Checkout(branchName string) bool {
	_, err := Command("checkout", "-b", branchName)
	if err != nil {
		_, err = Command("checkout", branchName)
	}
	return err == nil
}

// Stage stages the targeted files into git index.
func Stage(files []string) bool {
	if len(files) == 0 {
		return false
	}
	args := append([]string{"add"}, files...)
	_, err := Command(args...)
	return err == nil
}

// Commit records staged snapshots with a commit message.
func Commit(message string) bool {
	_, err := Command("commit", "-m", message)
	return err == nil
}

// ConfigureUser updates global git credentials.
func ConfigureUser(name, email string) bool {
	_, err1 := Command("config", "--global", "user.name", name)
	_, err2 := Command("config", "--global", "user.email", email)
	return err1 == nil && err2 == nil
}

// Init initializes a new git repository with the specified default branch.
func Init(defaultBranch string) bool {
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	_, err := Command("init", "-b", defaultBranch)
	if err != nil {
		// Fallback for older git versions that don't support `git init -b`
		_, errInit := Command("init")
		if errInit != nil {
			return false
		}
		_, _ = Command("branch", "-M", defaultBranch)
	}
	return true
}

// HasCommits checks if the repository contains at least one commit.
func HasCommits() bool {
	_, err := Command("rev-parse", "--verify", "HEAD")
	return err == nil
}

// RenameBranch renames the current active branch.
func RenameBranch(newBranch string) bool {
	if newBranch == "" {
		return false
	}
	_, err := Command("branch", "-M", newBranch)
	return err == nil
}

