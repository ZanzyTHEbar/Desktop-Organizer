package cmd

import (
	"desktop-cleaner/api"
	"desktop-cleaner/auth"
	"desktop-cleaner/internal"
	"desktop-cleaner/term"
	"fmt"

	"github.com/spf13/cobra"
)

var inviteCmd = &cobra.Command{
	Use:   "invite [email] [name] [org-role]",
	Short: "Invite a new user to the org",
	Run:   invite,
	Args:  cobra.MaximumNArgs(3),
}

func init() {
	RootCmd.AddCommand(inviteCmd)
}

func invite(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()

	email, name, orgRoleName := "", "", ""
	if len(args) >= 1 {
		email = args[0]
	}
	if len(args) >= 2 {
		name = args[1]
	}
	if len(args) == 3 {
		orgRoleName = args[2]
	}

	term.StartSpinner("")
	orgRoles, err := api.Client.ListOrgRoles()
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Failed to list org roles: %v", err)
	}

	if email == "" {
		var err error
		email, err = term.GetRequiredUserStringInput("Email:")
		if err != nil {
			term.OutputErrorAndExit("Failed to get email: %v", err)
		}
	}
	if name == "" {
		var err error
		name, err = term.GetRequiredUserStringInput("Name:")
		if err != nil {
			term.OutputErrorAndExit("Failed to get name: %v", err)
		}
	}

	if orgRoleName == "" {
		var orgRoleNames []string
		for _, orgRole := range orgRoles {
			orgRoleNames = append(orgRoleNames, orgRole.Label)
		}

		var err error
		orgRoleName, err = term.SelectFromList("Org role:", orgRoleNames)

		if err != nil {
			term.OutputErrorAndExit("Failed to select org role: %v", err)
		}
	}

	var orgRoleId string
	for _, orgRole := range orgRoles {
		if orgRole.Label == orgRoleName {
			orgRoleId = orgRole.Id
			break
		}
	}

	if orgRoleId == "" {
		term.OutputErrorAndExit("Org role '%s' not found", orgRoleName)
	}

	inviteRequest := internal.InviteRequest{
		Email:     email,
		Name:      name,
		OrgRoleId: orgRoleId,
	}

	term.StartSpinner("")
	apiErr := api.Client.InviteUser(inviteRequest)
	term.StopSpinner()
	if apiErr != nil {
		term.OutputErrorAndExit("Failed to invite user: %s", apiErr.Msg)
	}

	fmt.Println("✅ Invite sent")
}
