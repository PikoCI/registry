package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags for a type",
}

var tagsAddCmd = &cobra.Command{
	Use:   "add [namespace/name] [tag...]",
	Short: "Add tags to a type",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected format: namespace/name")
		}

		body := map[string]interface{}{
			"action": "add",
			"tags":   args[1:],
		}

		path := fmt.Sprintf("/api/plugins/%s/%s/tags", parts[0], parts[1])
		if err := client.doJSON("PATCH", path, body, nil); err != nil {
			return err
		}

		fmt.Printf("\u2713 Added tags to %s\n", args[0])
		return nil
	},
}

var tagsRemoveCmd = &cobra.Command{
	Use:   "remove [namespace/name] [tag...]",
	Short: "Remove tags from a type",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected format: namespace/name")
		}

		body := map[string]interface{}{
			"action": "remove",
			"tags":   args[1:],
		}

		path := fmt.Sprintf("/api/plugins/%s/%s/tags", parts[0], parts[1])
		if err := client.doJSON("PATCH", path, body, nil); err != nil {
			return err
		}

		fmt.Printf("\u2713 Removed tags from %s\n", args[0])
		return nil
	},
}

func init() {
	tagsCmd.AddCommand(tagsAddCmd)
	tagsCmd.AddCommand(tagsRemoveCmd)
}
