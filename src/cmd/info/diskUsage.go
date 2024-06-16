package info

import (
	"desktop-cleaner/auth"
	"desktop-cleaner/lib"
	"desktop-cleaner/term"

	"github.com/spf13/cobra"
)

var diskUsageCmd = &cobra.Command{
	Use:     "disk-usage",
	Aliases: []string{"du"},
	Short:   "Prints the disk usage of the current directory",
	Args:    cobra.MaximumNArgs(1),
	Run:     diskUsage,
}

func init() {

	InfoCmd.AddCommand(diskUsageCmd)
}

func diskUsage(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	if lib.CurrentPlanId == "" {
		term.OutputNoCurrentPlanErrorAndExit()
	}

	
}
