package cli

import (
	"desktop-cleaner/internal/cli"

	"github.com/spf13/cobra"
)

type HelpCMD struct {
	Help *cobra.Command
}

var helpShowAll bool

func NewHelpCMD(params *cli.CmdParams) *HelpCMD {
	return &HelpCMD{
		Help: NewHelp(params),
	}
}

func NewHelp(params *cli.CmdParams) *cobra.Command {
	helpCmd := &cobra.Command{
		Use:     "help",
		Aliases: []string{"h"},
		Short:   "Display help for DesktopCleaner",
		Long:    `Display help for DesktopCleaner.`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: params.term.PrintCustomHelp(helpShowAll)
		},
	}

	// add an --all/-a flag
	helpCmd.Flags().BoolVarP(&helpShowAll, "all", "a", false, "Show all commands")

	return helpCmd
}
