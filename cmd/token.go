package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens",
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		body := map[string]string{
			"name": args[0],
		}

		var result struct {
			Token string `json:"token"`
			Name  string `json:"name"`
		}

		if err := client.doJSON("POST", "/api/me/tokens", body, &result); err != nil {
			return err
		}

		fmt.Printf("\u2713 Token created: %s\n", result.Token)
		fmt.Println("Save this token - it won't be shown again.")
		return nil
	},
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your API tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		var results []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Prefix    string `json:"prefix"`
			CreatedAt string `json:"created_at"`
		}

		if err := client.doJSON("GET", "/api/me/tokens", nil, &results); err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No tokens found.")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "ID\tNAME\tPREFIX\tCREATED\n")
		for _, r := range results {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.ID, r.Name, r.Prefix, r.CreatedAt)
		}
		tw.Flush()

		return nil
	},
}

var tokenRevokeCmd = &cobra.Command{
	Use:   "revoke [token-id]",
	Short: "Revoke an API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		path := fmt.Sprintf("/api/me/tokens/%s", args[0])
		if err := client.doJSON("DELETE", path, nil, nil); err != nil {
			return err
		}

		fmt.Printf("\u2713 Token revoked\n")
		return nil
	},
}

func init() {
	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCmd.AddCommand(tokenListCmd)
	tokenCmd.AddCommand(tokenRevokeCmd)
}
