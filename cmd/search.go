package cmd

import (
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for types in the registry",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		params := url.Values{}
		if len(args) > 0 {
			params.Set("q", args[0])
		}
		if kind, _ := cmd.Flags().GetString("kind"); kind != "" {
			params.Set("kind", kind)
		}
		if tag, _ := cmd.Flags().GetString("tag"); tag != "" {
			params.Set("tag", tag)
		}

		var results []struct {
			Namespace   string `json:"namespace"`
			Name        string `json:"name"`
			Kind        string `json:"kind"`
			Description string `json:"description"`
			Downloads   int    `json:"downloads"`
		}

		if err := client.doJSON("GET", "/api/plugins?"+params.Encode(), nil, &results); err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "NAME\tKIND\tDESCRIPTION\tDOWNLOADS\n")
		for _, r := range results {
			desc := r.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Fprintf(tw, "%s/%s\t%s\t%s\t%d\n", r.Namespace, r.Name, r.Kind, desc, r.Downloads)
		}
		tw.Flush()

		return nil
	},
}

func init() {
	searchCmd.Flags().String("kind", "", "Filter by kind (resource_type, runner_type, etc.)")
	searchCmd.Flags().String("tag", "", "Filter by tag")
}
