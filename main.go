package main

import (
	"fmt"
	"os"

	"github.com/matthiasharzer/livestream-snapshotting-tool/cmd/run"
	"github.com/matthiasharzer/livestream-snapshotting-tool/cmd/version"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use: "livestream-snapshotting-tool",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCommand.AddCommand(version.Command)
	rootCommand.AddCommand(run.Command)
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
