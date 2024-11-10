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
	fmt.Printf("Terminal: %v", term)

	deskFS := deskfs.NewDesktopFS(term)

	fmt.Printf("DeskFS: %v", deskFS)

	// Setup the Root Command
	rootParams := &cli.CmdParams{
		Term:   term,
		DeskFS: deskFS,
	}

	fmt.Printf("RootParams: %v", rootParams)

	palette := generatePalette(rootParams)
	rootParams.Palette = palette

	rootCmd := cli.NewRootCMD(rootParams)

	fmt.Printf("RootCmd: %v", rootCmd)

	if err := rootCmd.Root.Execute(); err != nil {
		term.OutputErrorAndExit("Error executing root command: %v", err)
		slog.Error(fmt.Sprintf("Error executing root command: %v", err.Error()))
	}
}

func generatePalette(params *cli.CmdParams) []*cobra.Command {

	rewindCmd := git.NewRewind(params)
	rewind := cli.NewDesktopCleanerCMD(rewindCmd).Root
	helpUtil := cli.NewDesktopCleanerCMD(cli_util.NewHelp(params)).Root
	versionUtil := cli.NewDesktopCleanerCMD(cli_util.NewVersion(params)).Root
	upgradeUtil := cli.NewDesktopCleanerCMD(cli_util.NewUpgrade(params)).Root
	organize := cli.NewDesktopCleanerCMD(fs.NewOrganize(params)).Root
	//workspace := cli.NewDesktopCleanerCMD(workspace.NewWorkspace(params)).Root

	// Add commands here
	return []*cobra.Command{
		rewind,
		helpUtil,
		versionUtil,
		upgradeUtil,
		organize,
		//workspace,
	}
}
