package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var yankCmd = &cobra.Command{
	Use:   "yank [namespace/name] [version]",
	Short: "Yank a version from the registry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected format: namespace/name")
		}

		path := fmt.Sprintf("/api/plugins/%s/%s/%s/yank", parts[0], parts[1], args[1])
		if err := client.doJSON("POST", path, nil, nil); err != nil {
			return err
		}

		fmt.Printf("\u2713 Yanked %s/%s@%s\n", parts[0], parts[1], args[1])
		return nil
	},
}
