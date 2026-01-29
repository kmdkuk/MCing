package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kmdkuk/mcing/internal/cli/download"
	"github.com/kmdkuk/mcing/pkg/kube"
)

// NewDownloadCmd creates a new download command.
func NewDownloadCmd(opts *MCingOptions) *cobra.Command {
	o := download.NewOptions(opts.IOStreams, opts.ConfigFlags)
	cmd := &cobra.Command{
		Use:   "download <minecraft-name>",
		Short: "Download minecraft data directory",
		Long:  `Compress and download the /data directory of a specified Minecraft server to the local machine.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Namespace is explicitly required here because ConfigFlags logic handles it
			// lazily or needs the flags parsed. However, NewConfigFlags sets up flags.
			// The original logic relied on o.ConfigFlags to parse namespace later in Run.
			// Let's ensure flags are parsed by Cobra before RunE.
			if err := o.Complete(args); err != nil {
				return err
			}

			if o.Namespace == "" {
				var err error
				o.Namespace, _, err = o.ConfigFlags.ToRawKubeConfigLoader().Namespace()
				if err != nil {
					return err
				}
			}

			kubeExecutor := &kube.DefaultExecutor{
				Clientset:  opts.Clientset,
				RestConfig: opts.RestConfig,
			}

			d := download.NewDownloader(o, opts.K8sClient, kubeExecutor)
			return d.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Output, "output", "o", "", "Output filename (default: <minecraft-name>-data.tar.gz)")
	cmd.Flags().StringVarP(&o.Container, "container", "c", "minecraft", "Container name")

	return cmd
}
