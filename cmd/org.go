package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organizations",
}

var orgCreateCmd = &cobra.Command{
	Use:   "create [github-org]",
	Short: "Claim a namespace for a GitHub organization",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		body := map[string]string{
			"github_org": args[0],
		}

		if err := client.doJSON("POST", "/api/orgs", body, nil); err != nil {
			return err
		}

		fmt.Printf("\u2713 Claimed namespace for organization %s\n", args[0])
		return nil
	},
}

var orgInviteCmd = &cobra.Command{
	Use:   "invite [org] [username] [role]",
	Short: "Invite a member to an organization",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		body := map[string]string{
			"username": args[1],
			"role":     args[2],
		}

		path := fmt.Sprintf("/api/orgs/%s/members", args[0])
		if err := client.doJSON("POST", path, body, nil); err != nil {
			return err
		}

		fmt.Printf("\u2713 Invited %s to %s as %s\n", args[1], args[0], args[2])
		return nil
	},
}

func init() {
	orgCmd.AddCommand(orgCreateCmd)
	orgCmd.AddCommand(orgInviteCmd)
}
