package main

import (
	"desktop-cleaner/internal/cli"
	desktopFS "desktop-cleaner/internal/fs"
	"desktop-cleaner/internal/logger"
	"desktop-cleaner/internal/terminal"
	"log/slog"
)

func main() {
	// Setup the Dependancy Injection
	term := terminal.NewTerminal()
	deskFS := desktopFS.NewDesktopFS(term)
	logger.InitLogger(deskFS.InstanceConfig)

	// Setup the Root Command
	rootParams := &cli.CmdParams{
		Term:   term,
		DeskFS: deskFS,
	}

	rootCmd := cli.NewRootCMD(rootParams)

	if err := rootCmd.Root.Execute(); err != nil {
		term.OutputErrorAndExit("Error executing root command: %v", err)
		slog.Error("Error executing root command", err)
	}
}
