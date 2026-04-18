package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "unknown"

var Command = &cobra.Command{
	Use: "version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("livebuffer version %s\n", version)
	},
}
