package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kmdkuk/mcing/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var config struct {
	metricsAddr          string
	probeAddr            string
	enableLeaderElection bool
	zapOpts              zap.Options
	initImageName        string
	agentImageName       string
}

var rootCmd = &cobra.Command{
	Use:   "mcing-controller",
	Short: "mcing controller",
	Long:  "mcing controller",

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return subMain()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	fs := rootCmd.Flags()
	fs.StringVar(&config.metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&config.probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&config.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&config.initImageName, "init-image-name", "ghcr.io/kmdkuk/mcing-init:"+strings.TrimPrefix(version.Version, "v"), "mcing-init image name")
	fs.StringVar(&config.agentImageName, "agent-image-name", "ghcr.io/kmdkuk/mcing-agent:"+strings.TrimPrefix(version.Version, "v"), "mcing-agent image name")

	goflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(goflags)
	config.zapOpts.BindFlags(goflags)

	fs.AddGoFlagSet(goflags)
}
