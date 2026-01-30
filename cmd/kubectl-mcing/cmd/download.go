package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/kmdkuk/mcing/internal/cli/download"
	"github.com/kmdkuk/mcing/pkg/kube"
)

// NewDownloadCmd creates a new download command.
func NewDownloadCmd(opts *MCingOptions) *cobra.Command {
	o := download.NewOptions()
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
				o.Namespace, _, err = opts.ConfigFlags.ToRawKubeConfigLoader().Namespace()
				if err != nil {
					return err
				}
			}

			kubeExecutor := &kube.DefaultExecutor{
				Clientset:  opts.Clientset,
				RestConfig: opts.RestConfig,
			}

			d := download.NewDownloader(o, opts.K8sClient, kubeExecutor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt)
			go func() {
				<-sigCh
				cancel()
			}()
			return d.Run(ctx)
		},
	}

	cmd.Flags().StringVarP(&o.Output, "output", "o", "", "Output filename (default: <minecraft-name>-data.tar.gz)")

	return cmd
}
