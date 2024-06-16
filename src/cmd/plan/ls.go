package plan

import (
	"desktop-cleaner/api"
	"desktop-cleaner/auth"
	"desktop-cleaner/format"
	"desktop-cleaner/lib"
	"desktop-cleaner/term"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"ls"},
	Short:   "List everything in context",
	Run:     listContext,
}

func listContext(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	term.StartSpinner("")
	contexts, err := api.Client.ListContext(lib.CurrentPlanId, lib.CurrentBranch)
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error listing context: %v", err)
	}

	totalTokens := 0
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Name", "Type", "🪙", "Added", "Updated"})
	table.SetAutoWrapText(false)

	if len(contexts) == 0 {
		fmt.Println("🤷‍♂️ No context")
		fmt.Println()
		term.PrintCmds("", "load")
		return
	}

	for i, context := range contexts {
		totalTokens += context.NumTokens

		t, icon := context.TypeAndIcon()

		name := context.Name
		if len(name) > 40 {
			name = name[:20] + "⋯" + name[len(name)-20:]
		}

		row := []string{
			strconv.Itoa(i + 1),
			" " + icon + " " + name,
			t,
			strconv.Itoa(context.NumTokens), //+ " 🪙",
			format.Time(context.CreatedAt),
			format.Time(context.UpdatedAt),
		}
		table.Rich(row, []tablewriter.Colors{
			{tablewriter.Bold},
			{tablewriter.FgHiGreenColor, tablewriter.Bold},
		})
	}

	table.Render()

	tokensTbl := tablewriter.NewWriter(os.Stdout)
	tokensTbl.SetAutoWrapText(false)
	tokensTbl.Append([]string{color.New(term.ColorHiCyan, color.Bold).Sprintf("Total tokens →") + color.New(color.Bold).Sprintf(" %d 🪙", totalTokens)})

	tokensTbl.Render()

	fmt.Println()
	term.PrintCmds("", "load", "rm", "clear")

}

func init() {
	PlanCmd.AddCommand(contextCmd)
}
