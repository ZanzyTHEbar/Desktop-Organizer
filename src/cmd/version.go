package cmd

import (
	"fmt"

	"desktop-cleaner/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of DesktopCleaner",
	Long:  `All software has versions. This is DesktopCleaner's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
