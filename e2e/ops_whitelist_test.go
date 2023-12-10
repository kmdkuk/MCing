package e2e

import (
	_ "embed"
	"encoding/json"
	"path/filepath"

	"github.com/kmdkuk/mcing/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed testdata/add-ops-whitelist.yaml
var addOpsWhitelistYAML string

//go:embed testdata/no-ops-whitelist.yaml
var noOpsWhitelistYAML string

func testOpsWhitelist() {
	stsName := "mcing-ops-whitelist"
	It("should create no ops and whitelist", func() {
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
		})
	})
}
