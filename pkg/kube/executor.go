package kube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

// Executor abstracts Kubernetes operations.
type Executor interface {
	PortForward(namespace, podName string, remotePort int, out, errOut io.Writer) (int, chan struct{}, error)
	Exec(
		ctx context.Context,
		namespace, podName, container string,
		cmd []string,
		stdin io.Reader,
		out, errOut io.Writer,
	) error
}

// DefaultExecutor implements Executor.
type DefaultExecutor struct {
	Clientset  kubernetes.Interface
	RestConfig *rest.Config
}

// PortForward establishes a port forwarding session to a pod.
// It returns the local port, a stop channel to close the connection, and an error.
func (e *DefaultExecutor) PortForward(
	namespace string,
	podName string,
	remotePort int,
	outStream, errStream io.Writer,
) (int, chan struct{}, error) {
	req := e.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(e.RestConfig)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create round tripper: %w", err)
	}

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})
	// Use a strings.Builder if outStream is nil, as portforward requires a writer
	// but we might not capture it in the caller.
	// Actually, portforward.New requires a non-nil Writer for out.
	var out io.Writer
	if outStream == nil {
		out = new(strings.Builder)
	} else {
		out = outStream
	}

	var errOut io.Writer
	if errStream == nil {
		errOut = new(strings.Builder)
	} else {
		errOut = errStream
	}

	pf, err := portforward.New(
		spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL()),
		[]string{fmt.Sprintf("0:%d", remotePort)},
		stopCh,
		readyCh,
		out,
		errOut,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create port forwarder: %w", err)
	}

	go func() {
		if forwardErr := pf.ForwardPorts(); forwardErr != nil {
			// If provided, log to errStream. We can't really return this error easily since it's async.
			// The caller might see a closed channel or connection error.
			if errOut != nil {
				fmt.Fprintf(errOut, "PortForward error: %v\n", forwardErr)
			}
			close(stopCh)
		}
	}()

	<-readyCh
	ports, err := pf.GetPorts()
	if err != nil {
		close(stopCh)
		return 0, nil, fmt.Errorf("failed to get ports: %w", err)
	}
	if len(ports) == 0 {
		close(stopCh)
		return 0, nil, errors.New("failed to get forwarded ports")
	}

	return int(ports[0].Local), stopCh, nil
}

// Exec executes a command in a specific container of a pod.
func (e *DefaultExecutor) Exec(
	ctx context.Context,
	namespace string,
	podName string,
	containerName string,
	cmd []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) error {
	req := e.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(e.RestConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               false,
		TerminalSizeQueue: nil,
	})
}
