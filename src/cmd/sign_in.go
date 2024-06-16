package cmd

import (
	"desktop-cleaner/auth"
	"desktop-cleaner/term"

	"github.com/spf13/cobra"
)

var signInCmd = &cobra.Command{
	Use:   "sign-in",
	Short: "Sign in to a DesktopCleaner account",
	Args:  cobra.NoArgs,
	Run:   signIn,
}

func init() {
	RootCmd.AddCommand(signInCmd)
}

func signIn(cmd *cobra.Command, args []string) {
	err := auth.SelectOrSignInOrCreate()

	if err != nil {
		term.OutputErrorAndExit("Error signing in: %v", err)
	}
}
