package plan

import (
	"desktop-cleaner/api"
	"desktop-cleaner/auth"
	"desktop-cleaner/lib"
	"desktop-cleaner/term"
	"fmt"

	"desktop-cleaner/internal"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all context",
	Long:  `Clear all context.`,
	Run:   clearAllContext,
}

func clearAllContext(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	if lib.CurrentPlanId == "" {
		term.OutputNoCurrentPlanErrorAndExit()
	}

	term.StartSpinner("")
	contexts, err := api.Client.ListContext(lib.CurrentPlanId, lib.CurrentBranch)
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error retrieving context: %v", err)
	}

	deleteIds := map[string]bool{}

	for _, context := range contexts {
		deleteIds[context.Id] = true
	}

	if len(deleteIds) > 0 {
		res, err := api.Client.DeleteContext(lib.CurrentPlanId, lib.CurrentBranch, internal.DeleteContextRequest{
			Ids: deleteIds,
		})

		if err != nil {
			term.OutputErrorAndExit("Error deleting context: %v", err)
		}

		fmt.Println("✅ " + res.Msg)
	} else {
		fmt.Println("🤷‍♂️ No context removed")
	}

}

func init() {
	PlanCmd.AddCommand(clearCmd)
}
