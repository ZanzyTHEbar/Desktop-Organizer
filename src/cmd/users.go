package cmd

import (
	"desktop-cleaner/api"
	"desktop-cleaner/auth"
	"desktop-cleaner/internal"
	"desktop-cleaner/term"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users and pending invites and the current org",
	Run:   listUsersAndInvites,
}

func init() {
	RootCmd.AddCommand(usersCmd)
}

func listUsersAndInvites(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()

	var userResp *internal.ListUsersResponse
	var pendingInvites []*internal.Invite
	var orgRoles []*internal.OrgRole

	errCh := make(chan error)

	term.StartSpinner("")

	go func() {
		var err *internal.ApiError
		userResp, err = api.Client.ListUsers()
		if err != nil {
			errCh <- fmt.Errorf("error fetching users: %s", err.Msg)
			return
		}
		errCh <- nil
	}()

	go func() {
		var err *internal.ApiError
		pendingInvites, err = api.Client.ListPendingInvites()
		if err != nil {
			errCh <- fmt.Errorf("error fetching pending invites: %s", err.Msg)
			return
		}
		errCh <- nil
	}()

	go func() {
		var err *internal.ApiError
		orgRoles, err = api.Client.ListOrgRoles()
		if err != nil {
			errCh <- fmt.Errorf("error fetching org roles: %s", err.Msg)
			return
		}
		errCh <- nil

	}()

	for i := 0; i < 3; i++ {
		err := <-errCh
		if err != nil {
			term.StopSpinner()
			term.OutputErrorAndExit("%v", err)
		}
	}

	term.StopSpinner()

	orgRolesById := make(map[string]*internal.OrgRole)
	for _, role := range orgRoles {
		orgRolesById[role.Id] = role
	}

	// Display users and pending invites in a table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Email", "Name", "Role", "Status"})

	for _, user := range userResp.Users {
		table.Append([]string{user.Email, user.Name, orgRolesById[userResp.OrgUsersByUserId[user.Id].OrgRoleId].Label, "Active"})
	}

	for _, invite := range pendingInvites {
		table.Append([]string{invite.Email, invite.Name, orgRolesById[invite.OrgRoleId].Label, "Pending"})
	}

	table.Render()
}
