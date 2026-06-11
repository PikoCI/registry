package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [namespace]",
	Short: "List types in a namespace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		var results []struct {
			Name        string `json:"name"`
			Kind        string `json:"kind"`
			Description string `json:"description"`
			Downloads   int    `json:"downloads"`
		}

		if err := client.doJSON("GET", "/api/plugins/"+args[0], nil, &results); err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No types found.")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "NAME\tKIND\tDESCRIPTION\tDOWNLOADS\n")
		for _, r := range results {
			desc := r.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n", r.Name, r.Kind, desc, r.Downloads)
		}
		tw.Flush()

		return nil
	},
}
