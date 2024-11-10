package cli

import (
	"desktop-cleaner/internal/db"
	"desktop-cleaner/internal/deskfs"
	"desktop-cleaner/internal/terminal"

	"github.com/spf13/cobra"
)

type CmdParams struct {
	Term      *terminal.Terminal
	DeskFS    *deskfs.DesktopFS
	Palette   []*cobra.Command
	CentralDB *db.CentralDBProvider
}

type DesktopCleanerCMD struct {
	Root *cobra.Command
}

func NewDesktopCleanerCMD(cmdRoot *cobra.Command) *DesktopCleanerCMD {
	return &DesktopCleanerCMD{
		Root: cmdRoot,
	}
}
