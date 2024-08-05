package fs

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

// Initialize a Git repository if it doesn't exist
func InitGitRepo(directory string) error {
	gitDir := filepath.Join(directory, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = directory
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize git repository: %w", err)
		}
	}
	return nil
}

func IsGitRepo(dir string) bool {
	isGitRepo := false

	if isCommandAvailable("git") {
		// check whether we're in a git repo
		cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

		cmd.Dir = dir

		err := cmd.Run()

		if err == nil {
			isGitRepo = true
		}
	}

	return isGitRepo
}

func GitAddAndCommit(dir, message string, lockMutex bool) error {
	if lockMutex {
		gitMutex.Lock()
		defer gitMutex.Unlock()
	}

	err := GitAdd(dir, ".", false)
	if err != nil {
		return fmt.Errorf("error adding files to git repository for dir: %s, err: %v", dir, err)
	}

	err = GitCommit(dir, message, nil, false)
	if err != nil {
		return fmt.Errorf("error committing files to git repository for dir: %s, err: %v", dir, err)
	}

	return nil
}

func GitAddAndCommitPaths(dir, message string, paths []string, lockMutex bool) error {
	if len(paths) == 0 {
		return nil
	}

	if lockMutex {
		gitMutex.Lock()
		defer gitMutex.Unlock()
	}

	for _, path := range paths {
		err := GitAdd(dir, path, false)
		if err != nil {
			return fmt.Errorf("error adding file %s to git repository for dir: %s, err: %v", path, dir, err)
		}
	}

	err := GitCommit(dir, message, paths, false)
	if err != nil {
		return fmt.Errorf("error committing files to git repository for dir: %s, err: %v", dir, err)
	}

	return nil
}

func GitAdd(repoDir, path string, lockMutex bool) error {
	if lockMutex {
		gitMutex.Lock()
		defer gitMutex.Unlock()
	}

	res, err := exec.Command("git", "-C", repoDir, "add", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error adding files to git repository for dir: %s, err: %v, output: %s", repoDir, err, string(res))
	}

	return nil
}

func GitCommit(repoDir, commitMsg string, paths []string, lockMutex bool) error {
	if lockMutex {
		gitMutex.Lock()
		defer gitMutex.Unlock()
	}

	args := []string{"-C", repoDir, "commit", "-m", commitMsg, "--allow-empty"}

	if len(paths) > 0 {
		args = append(args, paths...)
	}

	res, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error committing files to git repository for dir: %s, err: %v, output: %s", repoDir, err, string(res))
	}

	return nil
}

func CheckUncommittedChanges() (bool, error) {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	// Check if there are any changes
	res, err := exec.Command("git", "status", "--porcelain").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("error checking for uncommitted changes: %v, output: %s", err, string(res))
	}

	// If there's output, there are uncommitted changes
	return strings.TrimSpace(string(res)) != "", nil
}

func GitStashCreate(message string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	res, err := exec.Command("git", "stash", "push", "--include-untracked", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating git stash: %v, output: %s", err, string(res))
	}

	return nil
}

// this matches output for git version 2.39.3
// need to test on other versions and check for more variations
// there isn't any structured way to get stash conflicts from git, unfortunately
const PopStashConflictMsg = "overwritten by merge"
const ConflictMsgFilesEnd = "commit your changes"

func GitStashPop(forceOverwrite bool) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	res, err := exec.Command("git", "stash", "pop").CombinedOutput()

	// we should no longer have conflicts since we are forcing an update before
	// running the 'apply' command as well as resetting any files with uncommitted change
	// still leaving this though in case something goes wrong

	if err != nil {
		log.Println("Error popping git stash:", string(res))

		if strings.Contains(string(res), PopStashConflictMsg) {
			log.Println("Conflicts detected")

			if !forceOverwrite {
				return fmt.Errorf("conflict popping git stash: %s", string(res))
			}

			// Parse the output to find which files have conflicts
			conflictFiles := parseConflictFiles(string(res))

			log.Println("Conflicting files:", conflictFiles)

			for _, file := range conflictFiles {
				// Reset each conflicting file individually
				checkoutRes, err := exec.Command("git", "checkout", "--ours", file).CombinedOutput()
				if err != nil {
					return fmt.Errorf("error resetting file %s: %v", file, string(checkoutRes))
				}
			}
			dropRes, err := exec.Command("git", "stash", "drop").CombinedOutput()
			if err != nil {
				return fmt.Errorf("error dropping git stash: %v", string(dropRes))
			}
			return nil
		} else {
			log.Println("No conflicts detected")

			return fmt.Errorf("error popping git stash: %v", string(res))
		}
	}

	return nil
}

func GitClearUncommittedChanges() error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	// Reset staged changes
	res, err := exec.Command("git", "reset", "--hard").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error resetting staged changes | err: %v, output: %s", err, string(res))
	}

	// Clean untracked files
	res, err = exec.Command("git", "clean", "-d", "-f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error cleaning untracked files | err: %v, output: %s", err, string(res))
	}

	return nil
}

func GitFileHasUncommittedChanges(path string) (bool, error) {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	res, err := exec.Command("git", "status", "--porcelain", path).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("error checking for uncommitted changes for file %s | err: %v, output: %s", path, err, string(res))
	}

	return strings.TrimSpace(string(res)) != "", nil
}

func GitCheckoutFile(path string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	res, err := exec.Command("git", "checkout", path).CombinedOutput()
	if err != nil {
		log.Println("Error checking out file:", string(res))

		return fmt.Errorf("error checking out file %s | err: %v, output: %s", path, err, string(res))
	}

	return nil
}

func GetCommitHistory() ([]string, error) {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	res, err := exec.Command("git", "rev-list", "--all").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting commit history | err: %v, output: %s", err, string(res))
	}

	return strings.Split(strings.TrimSpace(string(res)), "\n"), nil
}

func getTargetCommit(commits []string, steps int) (string, error) {
	if steps >= len(commits) && steps <= 0 && steps > 999 {
		return "", fmt.Errorf("invalid steps: exceeds number of available commits")
	}

	return commits[steps], nil
}

func rewindToCommit(targetSha string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	// Check if the provided sha exists in the logs
	commits, err := GetCommitHistory()
	if err != nil {
		return fmt.Errorf("error getting commit history: %v", err)
	}

	found := false

	for _, logSha := range commits {
		if logSha == targetSha {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("invalid sha: commit not found")
	}

	// Rewind to the specified Sha
	err = GitCheckoutFile(targetSha)
	if err != nil {
		return fmt.Errorf("error checking out file %s: %v", targetSha, err)
	}

	return nil
}

func isSHA(input string) bool {
	matched, _ := regexp.MatchString("^[a-f0-9]{40}$", input)
	return matched
}

func GitRewind(targetSha string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()

	if targetSha == "" {
		return fmt.Errorf("no target commit specified")
	}

	commits, err := GetCommitHistory()
	if err != nil {
		return fmt.Errorf("error getting commit history: %v", err)
	}

	var targetCommit string
	var msg string

	if isSHA(targetSha) {
		targetCommit = targetSha
		msg = "✅ Rewound to " + targetSha
	} else {
		numSteps, err := strconv.Atoi(targetSha)
		if err != nil {
			return fmt.Errorf("invalid target commit: %s", targetSha)
		}

		targetCommit, err = getTargetCommit(commits, numSteps)
		if err != nil {
			return fmt.Errorf("error getting target commit: %v", err)
		}

		postfix := "s"
		if numSteps == 1 {
			postfix = ""
		}

		msg = fmt.Sprintf("✅ Rewound %d step%s to %s", numSteps, postfix, targetSha)
	}

	err = rewindToCommit(targetCommit)
	if err != nil {
		return fmt.Errorf("error rewinding to %s: %v", targetCommit, err)
	}

	fmt.Println(msg)
	fmt.Println()

	return nil
}

func parseConflictFiles(gitOutput string) []string {
	var conflictFiles []string
	lines := strings.Split(gitOutput, "\n")

	inFilesSection := false

	for _, line := range lines {
		if inFilesSection {
			file := strings.TrimSpace(line)
			if file == "" {
				continue
			}
			conflictFiles = append(conflictFiles, strings.TrimSpace(line))
		} else if strings.Contains(line, PopStashConflictMsg) {
			inFilesSection = true
		} else if strings.Contains(line, ConflictMsgFilesEnd) {
			break
		}
	}
	return conflictFiles
}
