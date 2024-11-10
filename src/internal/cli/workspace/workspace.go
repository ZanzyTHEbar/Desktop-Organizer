package workspace

import (
	"desktop-cleaner/internal/cli"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type WorkspaceCMD struct {
	Workspace *cobra.Command
}

func NewWorkspace(params *cli.CmdParams) *cobra.Command {
	workspaceCmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   "Manage workspaces",
		Long:    `Manage workspaces including creating, updating, and deleting workspaces.`,
	}

	// Subcommand: create
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		Long:  `Create a new workspace with the specified root path and configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			rootPath, _ := cmd.Flags().GetString("root-path")
			config, _ := cmd.Flags().GetString("config")

			if rootPath == "" {
				params.Term.OutputWarning("Warn: root-path is required, using $(pwd)")
				var err error
				rootPath, err = os.Getwd()
				if err != nil {
					params.Term.OutputErrorAndExit("Error getting current working directory: %v", err)
				}
			}

			workspaceID, err := params.DeskFS.WorkspaceManager.CreateWorkspace(rootPath, config)
			if err != nil {
				params.Term.OutputErrorAndExit("Error creating workspace: %v", err)
			}
			params.Term.OutputSuccess(fmt.Sprintf("Workspace created successfully with ID: %d", workspaceID))
		},
	}
	createCmd.Flags().String("root-path", "", "Root path for the workspace (required)")
	createCmd.Flags().String("config", "", "Configuration data for the workspace")

	// Subcommand: update
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing workspace",
		Long:  `Update the configuration for an existing workspace by ID.`,
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := cmd.Flags().GetInt("id")
			config, _ := cmd.Flags().GetString("config")

			if id <= 0 {
				params.Term.OutputErrorAndExit("Error: valid workspace ID is required")
			}

			err := params.DeskFS.WorkspaceManager.UpdateWorkspace(id, config)
			if err != nil {
				params.Term.OutputErrorAndExit("Error updating workspace: %v", err)
			}
			params.Term.OutputSuccess(fmt.Sprintf("Workspace with ID %d updated successfully", id))
		},
	}
	updateCmd.Flags().Int("id", 0, "ID of the workspace to update (required)")
	updateCmd.Flags().String("config", "", "New configuration data for the workspace")

	// Subcommand: delete
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a workspace",
		Long:  `Delete an existing workspace by its ID.`,
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := cmd.Flags().GetInt("id")

			if id <= 0 {
				params.Term.OutputErrorAndExit("Error: valid workspace ID is required")
			}

			err := params.DeskFS.WorkspaceManager.DeleteWorkspace(id)
			if err != nil {
				params.Term.OutputErrorAndExit("Error deleting workspace: %v", err)
			}
			params.Term.OutputSuccess(fmt.Sprintf("Workspace with ID %d deleted successfully", id))
		},
	}
	deleteCmd.Flags().Int("id", 0, "ID of the workspace to delete (required)")

	// Add subcommands to the workspace command
	workspaceCmd.AddCommand(createCmd, updateCmd, deleteCmd)
	return workspaceCmd
}
