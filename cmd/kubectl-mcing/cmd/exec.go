package cmd

import (
	"context"
	"os"
	"time"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdexec "k8s.io/kubectl/pkg/cmd/exec"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var execConfig struct {
	stdin bool
	tty   bool
}

var execCmd = &cobra.Command{
	Use:  "exec MINECRAFT_NAME --- [COMMANDS]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRconCLI(cmd.Context(), args[0], cmd, args[1:])
	},
}

func runRconCLI(ctx context.Context, mcName string, cmd *cobra.Command, args []string) error {
	mc := &mcingv1alpha1.Minecraft{}
	err := kubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: mcName}, mc)
	if err != nil {
		return err
	}

	podName := getPodName(mc)
	commands := append([]string{podName, "--", "rcon-cli"}, args...)
	argsLenAtDash := 2
	options := &cmdexec.ExecOptions{
		StreamOptions: cmdexec.StreamOptions{
			IOStreams: genericclioptions.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stdout,
			},
			Stdin:         execConfig.stdin,
			TTY:           execConfig.tty,
			ContainerName: constants.MinecraftContainerName,
		},

		Executor: &cmdexec.DefaultRemoteExecutor{},
	}
	cmdutil.AddPodRunningTimeoutFlag(cmd, 3*time.Minute)
	cmdutil.CheckErr(options.Complete(factory, cmd, commands, argsLenAtDash))
	cmdutil.CheckErr(options.Validate())
	cmdutil.CheckErr(options.Run())

	return nil
}

func init() {
	fs := execCmd.Flags()
	fs.BoolVarP(&execConfig.stdin, "stdin", "i", false, "Pass stdin to the mysql container")
	fs.BoolVarP(&execConfig.tty, "tty", "t", false, "Allocate a TTY to stdin")

	rootCmd.AddCommand(execCmd)
}
