package e2e

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests
	corev1 "k8s.io/api/core/v1"

	"github.com/kmdkuk/mcing/pkg/config"
	"github.com/kmdkuk/mcing/pkg/constants"
)

//go:embed testdata/add-ops-whitelist.yaml.tmpl
var addOpsWhitelistYAML string

//go:embed testdata/no-ops-whitelist.yaml.tmpl
var noOpsWhitelistYAML string

//nolint:funlen // long test case
func testOpsWhitelist() {
	type userJSON struct {
		Name string `json:"name"`
	}

	It("should create no ops and whitelist", func() {
		name := "ops-whitelist-no-ops"
		data := map[string]any{
			"Name": name,
		}
		stsName := "mcing-" + name
		manifest := renderTemplate(noOpsWhitelistYAML, data)
		kubectlSafeWithInput(manifest, "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)

		Eventually(func(g Gomega) {
			stdout, _, err := kubectl(
				"exec",
				stsName+"-0",
				"--",
				"cat",
				filepath.Join(constants.DataPath, constants.OpsName),
			)
			g.Expect(err).ShouldNot(HaveOccurred())
			var ops []userJSON
			err = json.Unmarshal(stdout, &ops)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(ops).Should(BeEmpty())

			stdout, _, err = kubectl(
				"exec",
				stsName+"-0",
				"--",
				"cat",
				filepath.Join(constants.DataPath, constants.ServerPropsName),
			)
			g.Expect(err).ShouldNot(HaveOccurred())

			bf := bytes.NewBuffer(stdout)
			props, err := config.ParseServerProps(bf)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(props[constants.WhitelistProps]).Should(Equal("false"))
		}).Should(Succeed())

		kubectlSafeWithInput(manifest, "delete", "-f", "-")
	})

	It("should handle ops and whitelist lifecycle", func() {
		name := "ops-whitelist-lifecycle"
		data := map[string]any{
			"Name": name,
		}
		stsName := "mcing-" + name
		noOpsManifest := renderTemplate(noOpsWhitelistYAML, data)
		addOpsManifest := renderTemplate(addOpsWhitelistYAML, data)

		By("apply no ops and whitelist")
		kubectlSafeWithInput(noOpsManifest, "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)

		By("apply 1 ops and whitelist")
		kubectlSafeWithInput(addOpsManifest, "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl(
				"exec",
				stsName+"-0",
				"--",
				"cat",
				filepath.Join(constants.DataPath, constants.OpsName),
			)
			g.Expect(err).ShouldNot(HaveOccurred())
			type opsJSON struct {
				Name string `json:"name"`
			}
			var ops []opsJSON
			err = json.Unmarshal(stdout, &ops)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(ops).Should(HaveLen(1))
			g.Expect(ops[0].Name).Should(Equal("kmdkuk"))

			stdout, _, err = kubectl(
				"exec",
				stsName+"-0",
				"--",
				"cat",
				filepath.Join(constants.DataPath, constants.ServerPropsName),
			)
			g.Expect(err).ShouldNot(HaveOccurred())

			bf := bytes.NewBuffer(stdout)
			props, err := config.ParseServerProps(bf)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(props[constants.WhitelistProps]).Should(Equal("true"))

			stdout, _, err = kubectl(
				"exec",
				stsName+"-0",
				"--",
				"cat",
				filepath.Join(constants.DataPath, constants.WhiteListName),
			)
			g.Expect(err).ShouldNot(HaveOccurred())
			var whitelist []userJSON
			err = json.Unmarshal(stdout, &whitelist)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(whitelist).Should(HaveLen(1))
			g.Expect(whitelist[0].Name).Should(Equal("kmdkuk"))
		}).Should(Succeed())

		By("apply no ops and whitelist again")
		kubectlSafeWithInput(noOpsManifest, "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl(
				"exec",
				stsName+"-0",
				"--",
				"cat",
				filepath.Join(constants.DataPath, constants.OpsName),
			)
			g.Expect(err).ShouldNot(HaveOccurred())
			type opsJSON struct {
				Name string `json:"name"`
			}
			var ops []opsJSON
			err = json.Unmarshal(stdout, &ops)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(ops).Should(BeEmpty())

			stdout, _, err = kubectl(
				"exec",
				stsName+"-0",
				"--",
				"cat",
				filepath.Join(constants.DataPath, constants.ServerPropsName),
			)
			g.Expect(err).ShouldNot(HaveOccurred())

			bf := bytes.NewBuffer(stdout)
			props, err := config.ParseServerProps(bf)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(props[constants.WhitelistProps]).Should(Equal("false"))
		}).Should(Succeed())

		By("delete ops-whitelist instance")
		kubectlSafeWithInput(noOpsManifest, "delete", "-f", "-")
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl("get", "pod", "-o", "json", fmt.Sprintf("%s-0", stsName))
			g.Expect(err).ShouldNot(HaveOccurred())
			pods := &corev1.PodList{}
			err = json.Unmarshal(stdout, pods)
			g.Expect(err).ShouldNot(HaveOccurred())

			g.Expect(pods.Items).Should(BeEmpty())
		}).Should(Succeed())
	})
}
