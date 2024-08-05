package main

import (
	"desktop-cleaner/internal/cli"
	"desktop-cleaner/internal/cli/cli_util"
	"desktop-cleaner/internal/cli/git"
	desktopFS "desktop-cleaner/internal/fs"
	"desktop-cleaner/internal/terminal"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func main() {
	// Setup the Dependancy Injection
	term := terminal.NewTerminal()
	deskFS := desktopFS.NewDesktopFS(term)

	// Setup the Root Command
	rootParams := &cli.CmdParams{
		Term:   term,
		DeskFS: deskFS,
	}

	palette := generatePalette(rootParams)
	rootParams.Palette = palette

	rootCmd := cli.NewRootCMD(rootParams)

	if err := rootCmd.Root.Execute(); err != nil {
		term.OutputErrorAndExit("Error executing root command: %v", err)
		slog.Error(fmt.Sprintf("Error executing root command: %v", err.Error()))
	}
}

func generatePalette(params *cli.CmdParams) []*cobra.Command {
	rewind := git.NewRewindCMD(params)
	helpUtil := cli_util.NewHelpCMD(params)
	versionUtil := cli_util.NewVersionCMD(params)
	upgradeUtil := cli_util.NewUpgradeCMD(params)

	// Add commands here
	return []*cobra.Command{
		rewind.Rewind,
		helpUtil.Help,
		versionUtil.Version,
		upgradeUtil.Upgrade,
	}
}
