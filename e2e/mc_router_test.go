package e2e

import (
	_ "embed"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests
	corev1 "k8s.io/api/core/v1"

	"github.com/kmdkuk/mcing/pkg/constants"
)

//go:embed testdata/mc-router-minecraft.yaml.tmpl
var mcRouterMinecraftYAML string

func testMCRouter() {
	const (
		gatewayNS     = "mcing-gateway"
		testNS        = "default"
		defaultDomain = "minecraft.local"
	)

	It("should deploy mc-router gateway", func() {
		By("Verifying mc-router deployment exists")
		Eventually(func(g Gomega) {
			stdout, stderr, err := kubectl("get", "deployment", "mc-router", "-n", gatewayNS, "-o", "json")
			g.Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		}).Should(Succeed())

		By("Verifying mc-router service exists")
		Eventually(func(g Gomega) {
			stdout, stderr, err := kubectl("get", "service", "mc-router", "-n", gatewayNS, "-o", "json")
			g.Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		}).Should(Succeed())

		By("Waiting for mc-router deployment to be ready")
		waitDeployment(gatewayNS, "mc-router", 1)
	})

	It("should add mc-router annotation to Minecraft service", func() {
		name := "mc-router-test"
		stsName := "mcing-" + name
		data := map[string]any{
			"Name":      name,
			"Namespace": testNS,
		}
		manifest := renderTemplate(mcRouterMinecraftYAML, data)
		kubectlSafeWithInput(manifest, "apply", "-f", "-")

		defer func() {
			kubectlSafeWithInput(manifest, "delete", "-f", "-")
		}()

		waitStatefullSet(testNS, stsName, 1)

		By("Verifying service has mc-router annotation")
		Eventually(func(g Gomega) {
			stdout, stderr, err := kubectl("get", "service", stsName, "-n", testNS, "-o", "json")
			g.Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

			svc := &corev1.Service{}
			err = json.Unmarshal(stdout, svc)
			g.Expect(err).ShouldNot(HaveOccurred())

			annotation, ok := svc.Annotations[constants.MCRouterAnnotation]
			g.Expect(ok).Should(BeTrue(), "mc-router annotation should exist")
			expectedFQDN := fmt.Sprintf("%s.%s.%s", name, testNS, defaultDomain)
			g.Expect(annotation).Should(Equal(expectedFQDN))
		}).Should(Succeed())

		By("Verifying service is ClusterIP type")
		Eventually(func(g Gomega) {
			stdout, stderr, err := kubectl("get", "service", stsName, "-n", testNS, "-o", "json")
			g.Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

			svc := &corev1.Service{}
			err = json.Unmarshal(stdout, svc)
			g.Expect(err).ShouldNot(HaveOccurred())

			g.Expect(svc.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))
		}).Should(Succeed())
	})

	It("should use custom ExternalHostname when specified", func() {
		name := "custom-hostname"
		stsName := "mcing-" + name
		customHostname := "my-server.example.com"
		data := map[string]any{
			"Name":             name,
			"Namespace":        testNS,
			"ExternalHostname": customHostname,
		}
		manifest := renderTemplate(mcRouterMinecraftYAML, data)
		kubectlSafeWithInput(manifest, "apply", "-f", "-")

		defer func() {
			kubectlSafeWithInput(manifest, "delete", "-f", "-")
		}()

		waitStatefullSet(testNS, stsName, 1)

		By("Verifying service uses custom hostname in annotation")
		Eventually(func(g Gomega) {
			stdout, stderr, err := kubectl("get", "service", stsName, "-n", testNS, "-o", "json")
			g.Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

			svc := &corev1.Service{}
			err = json.Unmarshal(stdout, svc)
			g.Expect(err).ShouldNot(HaveOccurred())

			annotation := svc.Annotations[constants.MCRouterAnnotation]
			g.Expect(annotation).Should(Equal(customHostname))
		}).Should(Succeed())
	})
}
