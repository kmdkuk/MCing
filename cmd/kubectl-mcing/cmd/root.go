package cmd

import (
	"flag"
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	kubeConfigFlags *genericclioptions.ConfigFlags
	kubeClient      client.Client
	factory         util.Factory
	namespace       string
)

func init() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	kubeConfigFlags = genericclioptions.NewConfigFlags(true)
	kubeConfigFlags.AddFlags(rootCmd.PersistentFlags())
}

var rootCmd = &cobra.Command{
	Use:   "kubectl mcing",
	Short: "kubectl mcing",
	Long:  "kubectl mcing",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		factory = util.NewFactory(util.NewMatchVersionFlags(kubeConfigFlags))
		restConfig, err := factory.ToRESTConfig()
		if err != nil {
			return err
		}

		scheme := runtime.NewScheme()
		err = clientgoscheme.AddToScheme(scheme)
		if err != nil {
			return err
		}

		err = mcingv1alpha1.AddToScheme(scheme)
		if err != nil {
			return err
		}

		kubeClient, err = client.New(restConfig, client.Options{Scheme: scheme})
		if err != nil {
			return err
		}

		namespace, _, err = kubeConfigFlags.ToRawKubeConfigLoader().Namespace()
		return err
	},
}

func Execute() {
	defer klog.Flush()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
