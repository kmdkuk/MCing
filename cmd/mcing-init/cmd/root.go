package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Config represents the configuration for the init.
type Config struct {
	EnableLazyMC bool
}

// NewRootCmd represents the base command when called without any subcommands.
func NewRootCmd() *cobra.Command {
	var enableLazyMC bool
	rootCmd := &cobra.Command{
		Use:   "mcing-init",
		Short: "mcing init",
		Long:  "mcing init",

		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			cfg := Config{
				EnableLazyMC: enableLazyMC,
			}
			return subMain(cfg)
		},
	}

	fs := rootCmd.Flags()
	fs.BoolVar(&enableLazyMC, "enable-lazymc", false, "Enable LazyMC")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		//nolint:forbidigo // cli output
		fmt.Println(err)
		os.Exit(1)
	}
}
