package cmd

import (
	"desktop-cleaner/api"
	"desktop-cleaner/auth"
	"desktop-cleaner/format"
	"desktop-cleaner/internal"
	"desktop-cleaner/lib"
	"desktop-cleaner/term"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List plans with active or recently finished streams",
	Run:   ps,
}

func init() {
	RootCmd.AddCommand(psCmd)
}

func ps(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	if lib.CurrentPlanId == "" {
		term.OutputNoCurrentPlanErrorAndExit()
	}

	term.StartSpinner("")
	res, apiErr := api.Client.ListPlansRunning([]string{lib.CurrentProjectId}, true)
	term.StopSpinner()

	if apiErr != nil {
		term.OutputErrorAndExit("Error getting running plans: %v", apiErr)
		return
	}

	if len(res.Branches) == 0 {
		fmt.Println("🤷‍♂️ No active or recently finished streams")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Pid", "Plan", "Branch", "Started", "Status"})

	for _, b := range res.Branches {
		id := res.StreamIdByBranchId[b.Id]
		plan := res.PlansById[b.PlanId]

		status := "Active"
		finishedAt := res.StreamFinishedAtByBranchId[b.Id]
		switch b.Status {
		case internal.PlanStatusFinished:
			status = "Finished " + format.Time(finishedAt)
		case internal.PlanStatusError:
			status = "Error " + format.Time(finishedAt)
		case internal.PlanStatusStopped:
			status = "Stopped " + format.Time(finishedAt)
		case internal.PlanStatusMissingFile:
			status = "Missing file"
		}

		row := []string{
			id[:4],
			plan.Name,
			b.Name,
			format.Time(res.StreamStartedAtByBranchId[b.Id]),
			status,
		}

		var style []tablewriter.Colors
		if b.Name == lib.CurrentPlanId {
			style = []tablewriter.Colors{
				{tablewriter.FgGreenColor, tablewriter.Bold},
			}
		} else {
			style = []tablewriter.Colors{
				{tablewriter.Bold},
			}
		}

		table.Rich(row, style)

	}
	table.Render()

	fmt.Println()
	term.PrintCmds("", "connect", "stop")

}
