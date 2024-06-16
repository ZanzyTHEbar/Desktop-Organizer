package plan

import "github.com/spf13/cobra"

var PlanCmd = &cobra.Command{
	Use:     "plan",
	Aliases: []string{"f"},
	Short:   "Plan Operations",
	Long:    `Operations for managing plans.`,
	Run:     plan,
}

func plan(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {

}
