package cmd

import (
	"github.com/spf13/cobra"
)

var (
	AppName = "pikoci-registry"
	Version = "dev"
	Commit  = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: "PikoCI Registry - a versioned, searchable, tagged registry for PikoCI types",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = Version + " (" + Commit + ")"
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(yankCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(orgCmd)
	rootCmd.AddCommand(tokenCmd)
}

type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return "exit"
}
