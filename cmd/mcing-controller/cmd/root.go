package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kmdkuk/mcing/pkg/version"
)

// Config represents the configuration for the controller.
type Config struct {
	metricsAddr          string
	probeAddr            string
	enableLeaderElection bool
	webhookCertPath      string
	webhookCertName      string
	webhookCertKey       string
	zapOpts              zap.Options
	initImageName        string
	agentImageName       string
	interval             time.Duration
}

// NewRootCmd represents the base command when called without any subcommands.
func NewRootCmd() *cobra.Command {
	var (
		metricAddr           string
		probeAddr            string
		enableLeaderElection bool
		webhookCertPath      string
		webhookCertName      string
		webhookCertKey       string
		zapOpts              zap.Options
		initImageName        string
		agentImageName       string
		interval             time.Duration
	)

	rootCmd := &cobra.Command{
		Use:   "mcing-controller",
		Short: "mcing controller",
		Long:  "mcing controller",

		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			cfg := Config{
				metricsAddr:          metricAddr,
				probeAddr:            probeAddr,
				enableLeaderElection: enableLeaderElection,
				webhookCertPath:      webhookCertPath,
				webhookCertName:      webhookCertName,
				webhookCertKey:       webhookCertKey,
				zapOpts:              zapOpts,
				initImageName:        initImageName,
				agentImageName:       agentImageName,
				interval:             interval,
			}
			return subMain(cfg)
		},
	}

	fs := rootCmd.Flags()
	fs.StringVar(&metricAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(
		&webhookCertPath,
		"webhook-cert-path",
		"/tmp/k8s-webhook-server/serving-certs",
		"The directory that contains the webhook certificate.",
	)
	fs.StringVar(&webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	fs.StringVar(&webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	fs.StringVar(
		&initImageName,
		"init-image-name",
		"ghcr.io/kmdkuk/mcing-init:"+strings.TrimPrefix(version.Version, "v"),
		"mcing-init image name",
	)
	fs.StringVar(
		&agentImageName,
		"agent-image-name",
		"ghcr.io/kmdkuk/mcing-agent:"+strings.TrimPrefix(version.Version, "v"),
		"mcing-agent image name",
	)
	fs.DurationVar(&interval, "check-interval", 1*time.Minute, "Interval of minecraft maintenance")

	goflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(goflags)
	zapOpts.BindFlags(goflags)

	fs.AddGoFlagSet(goflags)

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
