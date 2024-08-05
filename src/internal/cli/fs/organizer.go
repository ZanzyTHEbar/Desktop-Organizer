package fs

import (
	"desktop-cleaner/internal/cli"
	deskFS "desktop-cleaner/internal/fs"
	"os"

	"github.com/spf13/cobra"
)

type OrganizeCMD struct {
	Organize *cobra.Command
}

var srcDir string
var targetDir string
var fileParams deskFS.FilePathParams

func NewOrganizeCMD(params *cli.CmdParams) *OrganizeCMD {
	return &OrganizeCMD{
		Organize: NewOrganize(params),
	}
}

func NewOrganize(params *cli.CmdParams) *cobra.Command {
	organizeCmd := &cobra.Command{
		Use:     "organize",
		Aliases: []string{"o"},
		Short:   "Organize files in the specified directory, based on the configuration",
		Long:    `Organize files based on the configuration. Optionally specify a destination directory. If not provided, the current working directory is used.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := organizeFiles(params); err != nil {
				params.Term.OutputErrorAndExit("Error organizing files: %v", err)
			}
		},
	}

	// Define flags and configuration settings
	organizeCmd.Flags().BoolVar(&fileParams.RemoveAfter, "remove", false, "Remove files after organizing")
	organizeCmd.Flags().BoolVar(&fileParams.NamesOnly, "names-only", false, "Organize by names only")
	organizeCmd.Flags().BoolVar(&fileParams.ForceSkipIgnore, "force-skip-ignore", false, "Force skip ignored files")
	organizeCmd.Flags().BoolVarP(&fileParams.Recursive, "recursive", "r", false, "Recursively organize files")
	organizeCmd.Flags().IntVarP(&fileParams.MaxDepth, "max-depth", "x", -1, "Maximum depth for recursion")
	organizeCmd.Flags().BoolVarP(&fileParams.GitEnabled, "git-enabled", "g", false, "Enable Git operations")
	organizeCmd.Flags().BoolVarP(&fileParams.CopyFiles, "copy", "c", false, "Enable move as Copy")
	organizeCmd.Flags().StringVarP(&srcDir, "srcDir", "d", "", "Destination directory to organize files from")
	organizeCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Target directory to organize files into")

	return organizeCmd
}

func organizeFiles(params *cli.CmdParams) error {

	// Set default directories if not provided
	if srcDir == "" {
		var err error
		srcDir, err = os.Getwd()
		if err != nil {
			params.Term.OutputErrorAndExit("Error getting current working directory: %v", err)
		}
	}

	if targetDir == "" {
		targetDir = srcDir
	}

	params.Term.ToggleSpinner(true, "Organizing files...")

	// Check if the destination directory is a Git repository
	if deskFS.IsGitRepo(srcDir) {
		params.Term.OutputInfo("Git repository detected.")
	} else {
		params.Term.OutputWarning("Git repository not detected. Proceeding without Git operations.")
	}

	// Perform the organization actions
	if err := params.DeskFS.EnhancedOrganize(srcDir, targetDir, *params.DeskFS.InstanceConfig, &fileParams); err != nil {
		params.Term.OutputErrorAndExit("Error organizing files: %v", err)
	}

	params.Term.ToggleSpinner(false, "")
	params.Term.OutputSuccess("Files organized successfully.")

	return nil
}
