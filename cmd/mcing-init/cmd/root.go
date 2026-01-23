package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCmd represents the base command when called without any subcommands.
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcing-init",
		Short: "mcing init",
		Long:  "mcing init",

		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			return subMain()
		},
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		//nolint:forbidigo // cli output
		fmt.Println(err)
		os.Exit(1)
	}
}
