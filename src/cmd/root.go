/*
Copyright © 2024 DaOfficialWizard pyr0ndet0s97@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"desktop-cleaner/cmd/fs"
	"desktop-cleaner/term"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var helpShowAll bool

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "desktop-cleaner [command] [flags]",
	Short: "DesktopCleaner is a tool to automate the clean up a specified directory",
	Run:   run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		term.OutputErrorAndExit("Error executing root command: %v", err)
		log.Fatalf("Error executing root command: %v", err)
	}
}

func run(cmd *cobra.Command, args []string) {
}

func addSubCommandPalettes() {
	RootCmd.AddCommand(fs.FSCmd)
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.desktop-cleaner/.desktop-cleaner.ini)")

	var helpCmd = &cobra.Command{
		Use:     "help",
		Aliases: []string{"h"},
		Short:   "Display help for DesktopCleaner",
		Long:    `Display help for DesktopCleaner.`,
		Run: func(cmd *cobra.Command, args []string) {
			term.PrintCustomHelp(helpShowAll)
		},
	}

	RootCmd.AddCommand(helpCmd)

	// add an --all/-a flag
	helpCmd.Flags().BoolVarP(&helpShowAll, "all", "a", false, "Show all commands")

	addSubCommandPalettes()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".desktop-cleaner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("ini")
		viper.SetConfigName(".desktop-cleaner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
