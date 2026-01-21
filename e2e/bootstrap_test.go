package e2e

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
)

func setupCluster() {
	By("delete all Minecraft")
	stdout := kubectlSafe("api-resources", "--api-group", "mcing.kmdkuk.com", "--no-headers")
	if strings.TrimSpace(string(stdout)) != "" {
		out, stderr, err := kubectl("get", "-A", "minecrafts.mcing.kmdkuk.com", "-o", "json")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", out, stderr, err)
		var ms mcingv1alpha1.MinecraftList
		err = json.Unmarshal(out, &ms)
		Expect(err).NotTo(HaveOccurred())
		for i := range ms.Items {
			m := &ms.Items[i]
			kubectlSafe("delete", "-n", m.Namespace, "minecrafts.mcing.kmdkuk.com", m.Name)
		}
		By("wait until all minecrafts are deleted")
		Eventually(func(g Gomega) {
			out, stderr, err := kubectl("get", "-A", "minecrafts.mcing.kmdkuk.com", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", out, stderr, err)
			var ms mcingv1alpha1.MinecraftList
			err = json.Unmarshal(out, &ms)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ms.Items).Should(BeEmpty())
		}).Should(Succeed())
	}

	By("delete namaspaces")
	_, _, err := kubectl("get", "ns", controllerNS)
	if err == nil {
		kubectlSafe("delete", "ns", controllerNS)
	}
	By("wait until namespace is deleted")
	Eventually(func(g Gomega) {
		out, stderr, getNSErr := kubectl("get", "ns", controllerNS, "-o", "json")
		g.Expect(getNSErr).To(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", out, stderr, getNSErr)
	}).Should(Succeed())

	By("create namespaces")
	createNamespace(controllerNS)

	By("deploy ../config/default for test")
	By("applying manifests")
	stdout, stderr, err := kustomizeBuild(".")
	Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	kubectlSafeWithInput(stdout, "apply", "-f", "-", "--server-side", "--force-conflicts")

	By("confirming all controller pods are ready")
	waitDeployment(controllerNS, "mcing-controller-manager", 1)

	// Wait until the webhook is ready by trying to create a dummy resource (dry-run).
	// This ensures that not only the Pod is running, but also the Service endpoints are propagated
	// and the CA bundle is properly injected by cert-manager.
	Eventually(func(g Gomega) {
		// Define a minimal CR for the dry-run check.
		stdout, stderr, err := kustomizeBuild("../config/samples")
		g.Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

		// Use --dry-run=server to check if the request passes through the webhook
		// without actually persisting the resource.
		// If the command succeeds, the webhook is reachable and working.
		stdout, stderr, err = kubectlWithInput(stdout, "create", "-f", "-", "--dry-run=server")
		g.Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	}).Should(Succeed())
}
