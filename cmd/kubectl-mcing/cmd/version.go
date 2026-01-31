package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/kmdkuk/mcing/pkg/version"
)

// NewVersionCmd creates a new version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubectl-mcing",
		Long:  `All software has versions. This is kubectl-mcing's`,
		Run: func(_ *cobra.Command, _ []string) {
			//nolint:forbidigo // Version information output
			fmt.Printf(
				"kubectl-mcing version: %s\n",
				version.Version,
			)
			//nolint:forbidigo // Version information output
			fmt.Printf(
				"Git revision: %s\n",
				version.Revision,
			)
			//nolint:forbidigo // Version information output
			fmt.Printf(
				"Build date: %s\n",
				version.BuildDate,
			)
			//nolint:forbidigo // Version information output
			fmt.Printf(
				"Go version: %s\n",
				runtime.Version(),
			)
			//nolint:forbidigo // Version information output
			fmt.Printf(
				"Go OS/Arch: %s/%s\n",
				runtime.GOOS,
				runtime.GOARCH,
			)
		},
	}
}
