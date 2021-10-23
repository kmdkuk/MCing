package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testBootstrap() {
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
