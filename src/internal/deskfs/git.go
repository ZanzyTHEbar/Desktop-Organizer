package deskfs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var gitMutex sync.Mutex

// Constants for detecting stash conflict messages
const (
	PopStashConflictMsg = "overwritten by merge"
	ConflictMsgFilesEnd = "commit your changes"
)

// Initialize a Git repository in the specified directory if it doesn't already exist.
func (dfs *DesktopFS) InitGitRepo(directory string) error {
	gitDir := filepath.Join(directory, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = directory
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize Git repository: %w", err)
		}
		log.Println("Initialized new Git repository at", directory)
	}
	return nil
}

// Check if the specified directory is a Git repository.
func (dfs *DesktopFS) IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// Stage all changes and commit them with the specified message.
func (dfs *DesktopFS) GitAddAndCommit(dir, message string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	if err := dfs.GitAdd(dir, "."); err != nil {
		return fmt.Errorf("error adding files to Git repository in dir %s: %w", dir, err)
	}

	if err := dfs.GitCommit(dir, message); err != nil {
		return fmt.Errorf("error committing files in Git repository in dir %s: %w", dir, err)
	}

	log.Println("Committed changes to Git with message:", message)
	return nil
}

// Stage specified paths in the repository.
func (dfs *DesktopFS) GitAdd(repoDir, path string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "add", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error adding files to Git repository at %s: %w | Output: %s", repoDir, err, string(output))
	}
	return nil
}

// Commit with a specified commit message.
func (dfs *DesktopFS) GitCommit(repoDir, commitMsg string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "commit", "-m", commitMsg, "--allow-empty")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error committing files in Git repository at %s: %w | Output: %s", repoDir, err, string(output))
	}
	return nil
}

// Check for uncommitted changes in the repository.
func (dfs *DesktopFS) CheckUncommittedChanges(repoDir string) (bool, error) {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("error checking for uncommitted changes: %w | Output: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// Stash all uncommitted changes with a specified message.
func (dfs *DesktopFS) GitStashCreate(repoDir, message string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "stash", "push", "--include-untracked", "-m", message)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error creating Git stash: %w | Output: %s", err, string(output))
	}
	return nil
}

// Pop the latest stash entry, resolving conflicts if specified.
func (dfs *DesktopFS) GitStashPop(repoDir string, forceOverwrite bool) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "stash", "pop")
	output, err := cmd.CombinedOutput()
	if err != nil && strings.Contains(string(output), PopStashConflictMsg) {
		log.Println("Conflicts detected while popping stash.")

		if forceOverwrite {
			conflictFiles := parseConflictFiles(string(output))
			for _, file := range conflictFiles {
				if resetErr := dfs.GitCheckoutFile(repoDir, file); resetErr != nil {
					return fmt.Errorf("error resolving conflict for file %s: %w", file, resetErr)
				}
			}
			exec.Command("git", "-C", repoDir, "stash", "drop").Run() // Drop the stash after resolution
			return nil
		}
		return fmt.Errorf("conflict encountered popping git stash: %s", output)
	} else if err != nil {
		return fmt.Errorf("error popping git stash: %w | Output: %s", err, string(output))
	}
	return nil
}

// Clears uncommitted changes, including untracked files.
func (dfs *DesktopFS) GitClearUncommittedChanges(repoDir string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	resetCmd := exec.Command("git", "-C", repoDir, "reset", "--hard")
	if output, err := resetCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error resetting changes: %w | Output: %s", err, string(output))
	}

	cleanCmd := exec.Command("git", "-C", repoDir, "clean", "-d", "-f")
	if output, err := cleanCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error cleaning untracked files: %w | Output: %s", err, string(output))
	}
	return nil
}

// Check if a specific file has uncommitted changes.
func (dfs *DesktopFS) GitFileHasUncommittedChanges(repoDir, path string) (bool, error) {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "status", "--porcelain", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("error checking uncommitted changes for file %s: %w | Output: %s", path, err, string(output))
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// Check out a file to discard local changes.
func (dfs *DesktopFS) GitCheckoutFile(repoDir, path string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "checkout", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error checking out file %s: %w | Output: %s", path, err, string(output))
	}
	return nil
}

// Retrieve commit history for the repository.
func (dfs *DesktopFS) GetCommitHistory(repoDir string) ([]string, error) {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	cmd := exec.Command("git", "-C", repoDir, "rev-list", "--all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error retrieving commit history: %w | Output: %s", err, string(output))
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// Rewind the repository to a specified commit.
func (dfs *DesktopFS) GitRewind(repoDir, targetSha string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	commits, err := dfs.GetCommitHistory(repoDir)
	if err != nil {
		return fmt.Errorf("error retrieving commit history: %w", err)
	}

	// Handle both SHA or step-based rewinding
	var targetCommit string
	if isSHA(targetSha) {
		targetCommit = targetSha
	} else {
		steps, err := strconv.Atoi(targetSha)
		if err != nil || steps < 0 || steps >= len(commits) {
			return fmt.Errorf("invalid target commit: %s", targetSha)
		}
		targetCommit = commits[steps]
	}

	cmd := exec.Command("git", "-C", repoDir, "checkout", targetCommit)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error rewinding to commit %s: %w | Output: %s", targetCommit, err, string(output))
	}
	return nil
}

func (dfs *DesktopFS) handleUncommittedChanges(dir string, params *FilePathParams) error {
	if params.GitEnabled {
		hasUncommitted, err := dfs.CheckUncommittedChanges(dir)
		if err != nil {
			return fmt.Errorf("error checking uncommitted changes: %w", err)
		}

		if hasUncommitted {
			fmt.Println("There are uncommitted changes. Would you like to stash them? (y/n)")
			var input string
			fmt.Scanln(&input)
			if strings.ToLower(input) == "y" {
				if err := dfs.GitStashCreate(dir, "Auto-stash before organizing"); err != nil {
					return fmt.Errorf("error creating git stash: %w", err)
				}
				fmt.Println("Changes stashed successfully.")
			}
		}
	}
	return nil
}

func (dfs *DesktopFS) clearChangesIfNeeded(dir string, params *FilePathParams) error {
	if params.GitEnabled {
		fmt.Println("Would you like to clear all uncommitted changes before organizing? (y/n)")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "y" {
			if err := dfs.GitClearUncommittedChanges(dir); err != nil {
				return fmt.Errorf("error clearing uncommitted changes: %w", err)
			}
			fmt.Println("All uncommitted changes have been cleared.")
		}
	}
	return nil
}

// Helper function to parse conflict files from Git output.
func parseConflictFiles(gitOutput string) []string {
	var conflictFiles []string
	lines := strings.Split(gitOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, PopStashConflictMsg) {
			conflictFiles = append(conflictFiles, strings.TrimSpace(line))
		} else if strings.Contains(line, ConflictMsgFilesEnd) {
			break
		}
	}
	return conflictFiles
}

// Utility to validate if a string is a valid SHA-1 hash.
func isSHA(input string) bool {
	matched, _ := regexp.MatchString("^[a-f0-9]{40}$", input)
	return matched
}
