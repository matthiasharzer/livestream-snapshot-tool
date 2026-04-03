package main

import (
	"fmt"
	"os"

	"github.com/matthiasharzer/livestream-snapshot-tool/cmd/version"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use: "livestream-snapshot-tool",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCommand.AddCommand(version.Command)
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
