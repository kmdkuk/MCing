package e2e

import (
	"encoding/json"
	"strings"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	. "github.com/onsi/ginkgo"
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
		kubectlSafeWithInput(stdout, "apply", "-f", "-")

		By("confirming all controller pods are ready")
		waitDeployment(controllerNS, "mcing-controller-manager", 1)
	})
}
