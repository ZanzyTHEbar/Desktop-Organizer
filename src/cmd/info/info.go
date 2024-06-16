package info

import "github.com/spf13/cobra"

var InfoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"f"},
	Short:   "Info Operations",
	Long:    `Operations for managing Infos.`,
	Run:     Info,
}

func Info(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {

}
