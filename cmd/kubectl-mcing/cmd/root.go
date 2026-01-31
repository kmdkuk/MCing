// Package cmd provides the commands for kubectl-mcing.
package cmd

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/version"
)

// MCingOptions holds configuration and clients for subcommands.
type MCingOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags
	IOStreams   genericclioptions.IOStreams

	RestConfig *rest.Config
	K8sClient  client.Client
	Clientset  *kubernetes.Clientset
}

// NewRootCmd creates a new root command.
func NewRootCmd() *cobra.Command {
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	o := &MCingOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
		RestConfig:  nil,
		K8sClient:   nil,
		Clientset:   nil,
	}

	rootCmd := &cobra.Command{
		Use:     "mcing",
		Version: version.Version,
		Short:   "kubectl plugin for mcing",
		Long: `mcing is a kubectl plugin for mcing.

This plugin provides commands to interact with Minecraft servers managed by MCing.
`,
		SilenceUsage: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			// Initialize klog flags
			klog.InitFlags(nil)
			var err error
			o.RestConfig, err = o.ConfigFlags.ToRESTConfig()
			if err != nil {
				return err
			}

			scheme := runtime.NewScheme()
			if err = mcingv1alpha1.AddToScheme(scheme); err != nil {
				return err
			}
			if err = corev1.AddToScheme(scheme); err != nil {
				return err
			}

			o.K8sClient, err = client.New(o.RestConfig, client.Options{Scheme: scheme})
			if err != nil {
				return err
			}

			o.Clientset, err = kubernetes.NewForConfig(o.RestConfig)
			if err != nil {
				return err
			}
			return nil
		},
	}

	o.ConfigFlags.AddFlags(rootCmd.PersistentFlags())
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	rootCmd.AddCommand(NewDownloadCmd(o))
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}
