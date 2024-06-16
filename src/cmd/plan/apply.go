package plan

import (
	"desktop-cleaner/auth"
	"desktop-cleaner/lib"
	"desktop-cleaner/term"

	"github.com/spf13/cobra"
)

var autoConfirm bool

var applyCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"ap"},
	Short:   "Apply a plan to the project",
	Args:    cobra.MaximumNArgs(1),
	Run:     apply,
}

func init() {
	applyCmd.Flags().BoolVarP(&autoConfirm, "yes", "y", false, "Automatically confirm unless plan is outdated")

	PlanCmd.AddCommand(applyCmd)
}

func apply(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	if lib.CurrentPlanId == "" {
		term.OutputNoCurrentPlanErrorAndExit()
	}

	lib.MustApplyPlan(lib.CurrentPlanId, lib.CurrentBranch, autoConfirm)
}
