package cli

import (
	desktopFS "desktop-cleaner/internal/fs"
	"desktop-cleaner/internal/terminal"

	"github.com/spf13/cobra"
)

type CmdParams struct {
	Term    *terminal.Terminal
	DeskFS  *desktopFS.DesktopFS
	Palette []*cobra.Command
}
