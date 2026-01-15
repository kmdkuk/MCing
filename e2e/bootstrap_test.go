package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testBootstrap() {
	It("delete all Minecraft", func() {
		stdout := kubectlSafe("api-resources", "--api-group", "mcing.kmdkuk.com", "--no-headers")
		if strings.TrimSpace(string(stdout)) == "" {
			Skip("mcing.kmdkuk.com does not exist")
		}
		stdout, stderr, err := kubectl("get", "-A", "minecrafts.mcing.kmdkuk.com", "-o", "json")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		var ms mcingv1alpha1.MinecraftList
		err = json.Unmarshal(stdout, &ms)
		Expect(err).NotTo(HaveOccurred())
		for i := range ms.Items {
			m := &ms.Items[i]
			kubectlSafe("delete", "-n", m.Namespace, "minecrafts.mcing.kmdkuk.com", m.Name)
		}
	})

	It("delete namaspaces", func() {
		_, _, err := kubectl("get", "ns", controllerNS)
		if err == nil {
			kubectlSafe("delete", "ns", controllerNS)
		}
	})

	It("create namespaces", func() {
		createNamespace(controllerNS)
	})

	It("should deploy ../config/default for test", func() {
		By("applying manifests")
		stdout, stderr, err := kustomizeBuild(".")
		Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		kubectlSafeWithInput(stdout, "apply", "-f", "-", "--server-side", "--force-conflicts")

		By("confirming all controller pods are ready")
		waitDeployment(controllerNS, "mcing-controller-manager", 1)

		// Wait until the webhook is ready by trying to create a dummy resource (dry-run).
		// This ensures that not only the Pod is running, but also the Service endpoints are propagated
		// and the CA bundle is properly injected by cert-manager.
		Eventually(func() error {
			// Define a minimal CR for the dry-run check.
			stdout, stderr, err := kustomizeBuild("../config/samples")
			Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			
			// Use --dry-run=server to check if the request passes through the webhook
			// without actually persisting the resource.
			_, stderr, err = kubectlWithInput(stdout, "create", "-f", "-", "--dry-run=server")

			// If the command succeeds, the webhook is reachable and working.
			if err == nil {
				return nil
			}

			// Return an error to retry if the connection is refused, or if CA/TLS errors occur.
			return fmt.Errorf("webhook not ready: %s", string(stderr))
		}, 10*time.Minute, 5*time.Second).Should(Succeed())
	})
}
