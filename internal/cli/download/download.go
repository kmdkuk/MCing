package download

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/kube"
	agent "github.com/kmdkuk/mcing/pkg/proto"
)

// AgentClientFactory is a function to create an agent client.
type AgentClientFactory func(port int) (agent.AgentClient, func() error, error)

func defaultAgentClientFactory(port int) (agent.AgentClient, func() error, error) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}
	return agent.NewAgentClient(conn), conn.Close, nil
}

// Options struct for holding download command options.
type Options struct {
	genericclioptions.IOStreams

	ConfigFlags *genericclioptions.ConfigFlags

	Namespace     string
	MinecraftName string
	Output        string
	Container     string
}

// NewOptions creates a new Options struct.
func NewOptions(
	streams genericclioptions.IOStreams,
	configFlags *genericclioptions.ConfigFlags,
) *Options {
	return &Options{
		ConfigFlags:   configFlags,
		IOStreams:     streams,
		Namespace:     "",
		MinecraftName: "",
		Output:        "",
		Container:     "",
	}
}

// Complete completes validation of the options.
func (o *Options) Complete(args []string) error {
	o.MinecraftName = args[0]

	if o.Output == "" {
		o.Output = fmt.Sprintf("%s-data.tar.gz", o.MinecraftName)
	}
	return nil
}

// Downloader struct for executing download logic.
type Downloader struct {
	Options *Options

	k8sClient    client.Client
	kubeExecutor kube.Executor
	agentFactory AgentClientFactory
}

// NewDownloader creates a new Downloader struct.
func NewDownloader(
	opts *Options,
	k8sClient client.Client,
	kubeExecutor kube.Executor,
) *Downloader {
	return &Downloader{
		Options:      opts,
		k8sClient:    k8sClient,
		kubeExecutor: kubeExecutor,
		agentFactory: defaultAgentClientFactory,
	}
}

// Run executes the download workflow.
func (d *Downloader) Run() error {
	var mc mcingv1alpha1.Minecraft
	err := d.k8sClient.Get(
		context.Background(),
		types.NamespacedName{Namespace: d.Options.Namespace, Name: d.Options.MinecraftName},
		&mc,
	)
	if err != nil {
		return fmt.Errorf("failed to get Minecraft resource: %w", err)
	}

	podName := mc.PodName()
	ctx := context.Background()

	cleanup, err := d.setupBackup(ctx, &mc)
	if err != nil {
		shouldSkip, checkErr := d.checkSleepingAndWarn(ctx, &mc, err)
		if checkErr != nil {
			return checkErr
		}
		if !shouldSkip {
			return err
		}
	} else {
		defer cleanup()
	}

	excludes := []string{"session.lock"}
	if mc.Spec.Backup.Excludes != nil {
		excludes = append(excludes, mc.Spec.Backup.Excludes...)
	}

	return d.performDownload(podName, excludes)
}

func (d *Downloader) setupBackup(ctx context.Context, mc *mcingv1alpha1.Minecraft) (func(), error) {
	podName := mc.PodName()

	// Port forward to agent
	localPort, stopCh, err := d.kubeExecutor.PortForward(
		d.Options.Namespace,
		podName,
		int(constants.AgentPort),
		nil,       // No stdout needed for portforward setup logs
		os.Stderr, // Log errors to stderr
	)
	if err != nil {
		return nil, err
	}

	agentClient, closeConn, err := d.agentFactory(localPort)
	if err != nil {
		close(stopCh)
		return nil, err
	}

	if err := d.prepareBackup(ctx, agentClient); err != nil {
		_ = closeConn()
		close(stopCh)
		return nil, err
	}

	return func() {
		// New context for cleanup as main context might be cancelled
		cleanupCtx := context.Background()
		if err := d.cleanupBackup(cleanupCtx, agentClient); err != nil {
			klog.Errorf("Failed to execute save-on: %v", err)
		}
		_ = closeConn()
		close(stopCh)
	}, nil
}

func (d *Downloader) checkSleepingAndWarn(
	ctx context.Context,
	mc *mcingv1alpha1.Minecraft,
	originalErr error,
) (bool, error) {
	isSleeping, err := d.isServerSleeping(ctx, mc)
	if err != nil {
		// Failed to check sleep status, return original error combined with check error
		return false, fmt.Errorf("%w (also failed to check sleep status: %w)", originalErr, err)
	}
	if isSleeping {
		klog.Warningf(
			"Server appears to be sleeping (AutoPause enabled). Skipping backup preparation (save-off/save-all). Original error: %v",
			originalErr,
		)
		return true, nil
	}
	return false, originalErr
}

// isServerSleeping checks if the Minecraft server is sleeping loop.
// TODO: This check should be performed by the controller and reflected in the resource status.
// For example, Status.Phase should indicate "Sleeping" or similar.
// Then, this CLI command can just check the status instead of executing pgrep.
func (d *Downloader) isServerSleeping(ctx context.Context, mc *mcingv1alpha1.Minecraft) (bool, error) {
	if mc.Spec.AutoPause.Enabled != nil && !*mc.Spec.AutoPause.Enabled {
		return false, nil
	}

	var stdout bytes.Buffer
	// Check if java process exists. "pgrep java" exit code 1 means no process found.
	err := d.kubeExecutor.Exec(
		ctx,
		d.Options.Namespace,
		mc.PodName(),
		d.Options.Container,
		[]string{"pgrep", "java"},
		nil,
		&stdout,
		nil,
	)
	if err != nil {
		// If command terminated with exit code 1, it implies process not found -> Sleeping
		if strings.Contains(err.Error(), "exit code 1") {
			return true, nil
		}
		return false, err
	}

	// If err is nil, process found -> Not sleeping
	return false, nil
}

func (d *Downloader) prepareBackup(ctx context.Context, client agent.AgentClient) error {
	klog.Info("Disabling auto-save...")
	if _, err := client.SaveOff(ctx, &agent.SaveOffRequest{}); err != nil {
		return fmt.Errorf("failed to execute save-off: %w", err)
	}

	klog.Info("Saving game to disk...")
	if _, err := client.SaveAll(ctx, &agent.SaveAllRequest{}); err != nil {
		return fmt.Errorf("failed to execute save-all: %w", err)
	}
	return nil
}

func (d *Downloader) cleanupBackup(ctx context.Context, client agent.AgentClient) error {
	klog.Info("Enabling auto-save...")
	_, err := client.SaveOn(ctx, &agent.SaveOnRequest{})
	return err
}

func (d *Downloader) performDownload(podName string, excludes []string) error {
	klog.Infof("Downloading data to %s...", d.Options.Output)
	tarCmd := []string{"tar", "czf", "-", "-C", "/data"}
	for _, ex := range excludes {
		tarCmd = append(tarCmd, "--exclude", ex)
	}
	tarCmd = append(tarCmd, ".")

	outFile, err := os.Create(d.Options.Output)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Handle interrupt signal to cancel the stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := d.kubeExecutor.Exec(
		ctx,
		d.Options.Namespace,
		podName,
		d.Options.Container,
		tarCmd,
		nil,
		outFile,
		os.Stderr,
	); err != nil {
		return fmt.Errorf("failed to download data: %w", err)
	}

	klog.Info("Download completed successfully.")
	return nil
}
