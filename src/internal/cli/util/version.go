package cli

import (
	"desktop-cleaner/internal/cli"
	"desktop-cleaner/version"
	"fmt"

	"github.com/spf13/cobra"
)

type VersionCMD struct {
	Version *cobra.Command
}

func NewVersionCMD(params *cli.CmdParams) *VersionCMD {
	return &VersionCMD{
		Version: NewVersion(params),
	}
}

func NewVersion(params *cli.CmdParams) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of DesktopCleaner",
		Long:  `All software has versions. This is DesktopCleaner's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version)
		},
	}

	return versionCmd
}
