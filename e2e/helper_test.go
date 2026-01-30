package e2e

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/gomega" //nolint:revive // dot imports for tests
	appsv1 "k8s.io/api/apps/v1"
)

//go:embed testdata/minecraft-container.yaml.tmpl
var minecraftContainerTmpl string

//go:embed testdata/minecraft.yaml.tmpl
var minecraftTmpl string

func renderTemplate(tmplStr string, data any) []byte {
	tmpl, err := template.New("manifest").Parse(tmplStr)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	_, err = tmpl.New("minecraft-container").Parse(minecraftContainerTmpl)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return buf.Bytes()
}

func kubectl(args ...string) ([]byte, []byte, error) {
	return execAtLocal("kubectl", nil, args...)
}

func kubectlWithInput(input []byte, args ...string) ([]byte, []byte, error) {
	return execAtLocal("kubectl", input, args...)
}

func kubectlSafe(args ...string) []byte {
	stdout, stderr, err := kubectl(args...)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	return stdout
}

func kubectlSafeWithInput(input []byte, args ...string) []byte {
	stdout, stderr, err := kubectlWithInput(input, args...)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	return stdout
}

func kubectlMcing(args ...string) ([]byte, []byte, error) {
	return execAtLocal(filepath.Join(binDir, "kubectl-mcing"), nil, args...)
}

func kubectlMcingSafe(args ...string) ([]byte, []byte) {
	stdout, stderr, err := kubectlMcing(args...)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	return stdout, stderr
}

func kustomizeBuild(dir string) ([]byte, []byte, error) {
	return execAtLocal("kustomize", nil, "build", dir)
}

func execAtLocal(cmd string, input []byte, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	command := exec.CommandContext(context.Background(), cmd, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	if len(input) != 0 {
		command.Stdin = bytes.NewReader(input)
	}

	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func createNamespace(ns string) {
	stdout, stderr, err := kubectl("create", "namespace", ns)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

	EventuallyWithOffset(1, func() error {
		stdout, stderr, err := kubectl("get", "sa", "default", "-n", ns)
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %w", stdout, stderr, err)
		}
		return nil
	}).Should(Succeed())
}

func waitDeployment(namespace, name string, replicas int) {
	EventuallyWithOffset(1, func() error {
		stdout, stderr, err := kubectl("get", "deployment", name, "-n", namespace, "-o", "json")
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %w", stdout, stderr, err)
		}

		d := new(appsv1.Deployment)
		err = json.Unmarshal(stdout, d)
		if err != nil {
			return err
		}

		if int(d.Status.AvailableReplicas) != replicas {
			return fmt.Errorf("AvailableReplicas is not %d: %d", replicas, int(d.Status.AvailableReplicas))
		}

		return nil
	}).ShouldNot(HaveOccurred())
}

func waitStatefullSet(namespace, name string, replicas int) { //nolint:unparam // replicas is always 1 in current tests
	EventuallyWithOffset(1, func() error {
		stdout, stderr, err := kubectl("get", "statefulset", name, "-n", namespace, "-o", "json")
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %w", stdout, stderr, err)
		}

		d := new(appsv1.StatefulSet)
		err = json.Unmarshal(stdout, d)
		if err != nil {
			return err
		}

		if int(d.Status.AvailableReplicas) != replicas {
			return fmt.Errorf("AvailableReplicas is not %d: %d", replicas, int(d.Status.AvailableReplicas))
		}

		return nil
	}).ShouldNot(HaveOccurred())
}

//nolint:nonamedreturns // required to set err in defer
func getFreePort() (port int, err error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return -1, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1, err
	}
	defer func() {
		closeErr := l.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return -1, errors.New("failed to get TCP address")
	}
	port = tcpAddr.Port
	return port, nil
}

func PortForwardCmd(ctx context.Context, namespace, name string, to int) (*exec.Cmd, int, error) {
	localPort, err := getFreePort()
	if err != nil {
		return nil, -1, err
	}
	return exec.CommandContext( //nolint:gosec // for test code
		ctx,
		"kubectl",
		"port-forward",
		"-n",
		namespace,
		name,
		fmt.Sprintf("%d:%d", localPort, to),
	), localPort, nil
}

func verifyArchive(tarPath, expectedFile, expectedContent string) {
	// Verify that the tar archive contains the expected file with expected content.
	// We use "tar -tf" to list and "tar -xf ... -O" to extract to stdout.

	// Check if file exists in archive
	stdout, _, err := execAtLocal("tar", nil, "-tf", tarPath)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to list tar contents")
	ExpectWithOffset(
		1,
		string(stdout),
	).To(ContainSubstring(expectedFile), "archive does not contain expected file: %s", expectedFile)

	// Check content
	stdout, _, err = execAtLocal("tar", nil, "-xf", tarPath, expectedFile, "-O")
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to extract file from tar")
	ExpectWithOffset(1, string(stdout)).To(ContainSubstring(expectedContent), "file content does not match")
}
