package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
)

func kubectl(args ...string) ([]byte, []byte, error) {
	return execAtLocal(filepath.Join(binDir, "kubectl"), nil, args...)
}

func kubectlWithInput(input []byte, args ...string) ([]byte, []byte, error) {
	return execAtLocal(filepath.Join(binDir, "kubectl"), input, args...)
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

func kustomizeBuild(dir string) ([]byte, []byte, error) {
	return execAtLocal(filepath.Join(binDir, "kustomize"), nil, "build", dir)
}

func execAtLocal(cmd string, input []byte, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command(cmd, args...)
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
			return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		return nil
	}).Should(Succeed())
}

func waitDeployment(namespace, name string, replicas int) {
	EventuallyWithOffset(1, func() error {
		stdout, stderr, err := kubectl("get", "deployment", name, "-n", namespace, "-o", "json")
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
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

func waitStatefullSet(namespace, name string, replicas int) {
	EventuallyWithOffset(1, func() error {
		stdout, stderr, err := kubectl("get", "statefulset", name, "-n", namespace, "-o", "json")
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
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
