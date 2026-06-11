package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [namespace/name] [version]",
	Short: "Show type or version details",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected format: namespace/name")
		}

		path := fmt.Sprintf("/api/plugins/%s/%s", parts[0], parts[1])
		if len(args) == 2 {
			path = fmt.Sprintf("/api/plugins/%s/%s/%s", parts[0], parts[1], args[1])
		}

		var result map[string]interface{}
		if err := client.doJSON("GET", path, nil, &result); err != nil {
			return err
		}

		for k, v := range result {
			fmt.Printf("%-15s %v\n", k+":", v)
		}

		return nil
	},
}
