package main

import (
	"fmt"
	"os"

	"template/cmd/version"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use: "<tool-name>",
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
