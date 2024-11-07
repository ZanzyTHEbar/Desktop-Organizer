package main

import (
	"desktop-cleaner/internal/cli"
	"desktop-cleaner/internal/cli/cli_util"
	"desktop-cleaner/internal/cli/fs"
	"desktop-cleaner/internal/cli/git"
	"desktop-cleaner/internal/deskfs"
	"desktop-cleaner/internal/terminal"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func main() {
	// Setup the Dependancy Injection
	term := terminal.NewTerminal()
	deskFS := deskfs.NewDesktopFS(term)

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
	rewind := cli.NewDesktopCleanerCMD(git.NewRewind(params)).Root
	helpUtil := cli.NewDesktopCleanerCMD(cli_util.NewHelp(params)).Root
	versionUtil := cli.NewDesktopCleanerCMD(cli_util.NewVersion(params)).Root
	upgradeUtil := cli.NewDesktopCleanerCMD(cli_util.NewUpgrade(params)).Root
	organize := cli.NewDesktopCleanerCMD(fs.NewOrganize(params)).Root

	// Add commands here
	return []*cobra.Command{
		rewind,
		helpUtil,
		versionUtil,
		upgradeUtil,
		organize,
	}
}
