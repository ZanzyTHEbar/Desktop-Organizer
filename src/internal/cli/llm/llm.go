package llm

import (
	"desktop-cleaner/internal/cli"
	"fmt"

	"github.com/spf13/cobra"
)

type LLMAgentCMD struct {
	LLMAgent *cobra.Command
}

func NewRewind(params *cli.CmdParams) *cobra.Command {
	llmagentCmd := &cobra.Command{
		Use:     "llmagent [prompt]",
		Aliases: []string{"llm"},
		Short:   "Run the LLM agent to reorganize files in the specified directory",
		Long: `Run the LLM agent to reorganize files based on the configuration. Optionally specify a prompt. If not provided, the default prompt is used.
		
		You can pass a "prompt" string to the LLM agent to generate recommendations based on the prompt. If no prompt is passed, the default prompt will be used.

		Example:

		$ desktop-cleaner llm "Organize files from this week in my work directory based on relevance, create subfolders for each well defined category."`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			llmagent(params, args)
		},
	}

	return llmagentCmd
}

func llmagent(params *cli.CmdParams, args []string) {
	var stepsOrSha string
	if len(args) > 0 {
		stepsOrSha = args[0]
	} else {
		stepsOrSha = "1"
	}

	params.Term.ToggleSpinner(true, fmt.Sprintf("Rewinding to %s ...", stepsOrSha))

	// TODO: Implement LLM agent logic here

	params.Term.ToggleSpinner(false, "")
}
