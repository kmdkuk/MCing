package e2e

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"path/filepath"

	"github.com/kmdkuk/mcing/pkg/config"
	"github.com/kmdkuk/mcing/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

//go:embed testdata/add-ops-whitelist.yaml
var addOpsWhitelistYAML string

//go:embed testdata/no-ops-whitelist.yaml
var noOpsWhitelistYAML string

func testOpsWhitelist() {
	stsName := "mcing-ops-whitelist"
	type userJson struct {
		Name string `json:"name"`
	}
	It("should create no ops and whitelist", func() {
		kubectlSafeWithInput([]byte(noOpsWhitelistYAML), "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl("exec", stsName+"-0", "--", "cat", filepath.Join(constants.DataPath, constants.OpsName))
			g.Expect(err).ShouldNot(HaveOccurred())
			var ops []userJson
			err = json.Unmarshal(stdout, &ops)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(ops).Should(HaveLen(0))

			stdout, _, err = kubectl("exec", stsName+"-0", "--", "cat", filepath.Join(constants.DataPath, constants.ServerPropsName))
			g.Expect(err).ShouldNot(HaveOccurred())

			bf := bytes.NewBuffer(stdout)
			props := config.ParseServerProps(bf)
			g.Expect(props[constants.WhitelistProps]).Should(Equal("false"))
		})
	})

	It("should apply 1 ops and whitelist", func() {
		kubectlSafeWithInput([]byte(addOpsWhitelistYAML), "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl("exec", stsName+"-0", "--", "cat", filepath.Join(constants.DataPath, constants.OpsName))
			g.Expect(err).ShouldNot(HaveOccurred())
			type opsJson struct {
				Name string `json:"name"`
			}
			var ops []opsJson
			err = json.Unmarshal(stdout, &ops)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(ops).Should(HaveLen(1))
			g.Expect(ops[0].Name).Should(Equal("kmdkuk"))

			stdout, _, err = kubectl("exec", stsName+"-0", "--", "cat", filepath.Join(constants.DataPath, constants.ServerPropsName))
			g.Expect(err).ShouldNot(HaveOccurred())

			bf := bytes.NewBuffer(stdout)
			props := config.ParseServerProps(bf)
			g.Expect(props[constants.WhitelistProps]).Should(Equal("true"))

			stdout, _, err = kubectl("exec", stsName+"-0", "--", "cat", filepath.Join(constants.DataPath, constants.WhiteListName))
			g.Expect(err).ShouldNot(HaveOccurred())
			var whitelist []userJson
			err = json.Unmarshal(stdout, &ops)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(whitelist).Should(HaveLen(1))
			g.Expect(whitelist[0].Name).Should(Equal("kmdkuk"))
		})
	})

	It("should apply no ops and whitelist", func() {
		kubectlSafeWithInput([]byte(noOpsWhitelistYAML), "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl("exec", stsName+"-0", "--", "cat", filepath.Join(constants.DataPath, constants.OpsName))
			g.Expect(err).ShouldNot(HaveOccurred())
			type opsJson struct {
				Name string `json:"name"`
			}
			var ops []opsJson
			err = json.Unmarshal(stdout, &ops)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(ops).Should(HaveLen(0))

			stdout, _, err = kubectl("exec", stsName+"-0", "--", "cat", filepath.Join(constants.DataPath, constants.ServerPropsName))
			g.Expect(err).ShouldNot(HaveOccurred())

			bf := bytes.NewBuffer(stdout)
			props := config.ParseServerProps(bf)
			g.Expect(props[constants.WhitelistProps]).Should(Equal("false"))
		})
	})

	It("should delete ops-whitelist instance", func() {
		kubectlSafeWithInput([]byte(noOpsWhitelistYAML), "delete", "-f", "-")
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl("get", "pod", "-o", "json")
			g.Expect(err).ShouldNot(HaveOccurred())
			pods := &corev1.PodList{}
			err = json.Unmarshal(stdout, pods)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(pods).Should(HaveLen(0))
		})
	})
}
