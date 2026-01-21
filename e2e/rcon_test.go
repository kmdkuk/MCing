package e2e

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests
)

//go:embed testdata/minecraft-rcon.yaml.tmpl
var minecraftRconTemplate string

//go:embed testdata/secret-rcon.yaml.tmpl
var secretRconTemplate string

// function moved to suite_test.go

func testRcon() {
	Describe("RCON Password", func() {
		It("should work with auto-generated secret", func() {
			name := "rcon-auto"
			// 1. Create Minecraft CR
			data := map[string]any{
				"Name": name,
			}
			manifest := renderTemplate(minecraftRconTemplate, data)
			kubectlSafeWithInput(manifest, "apply", "-f", "-")

			// 2. Wait for Pod Ready
			waitStatefullSet("default", "mcing-"+name, 1)

			// 3. Get generated Secret
			out, _, err := kubectl(
				"get",
				"secret",
				"mcing-"+name+"-rcon-password",
				"-o",
				"jsonpath={.data.rcon-password}",
			)
			Expect(err).NotTo(HaveOccurred())
			passwordBytes, err := base64.StdEncoding.DecodeString(string(out))
			Expect(err).NotTo(HaveOccurred())
			password := string(passwordBytes)
			Expect(password).NotTo(BeEmpty())

			// 4. Verify RCON connection
			// rcon-cli needs some time to be ready after server start
			Eventually(func() error {
				_, stderr, err := kubectl(
					"exec",
					"mcing-"+name+"-0",
					"-c",
					"minecraft",
					"--",
					"rcon-cli",
					"--password",
					password,
					"list",
				)
				if err != nil {
					return fmt.Errorf("err: %w, stderr: %s", err, stderr)
				}
				return nil
			}).Should(Succeed())

			// Cleanup
			kubectlSafeWithInput(manifest, "delete", "-f", "-")
		})

		It("should work with user-specified secret", func() {
			name := "rcon-custom"
			secretName := "my-rcon-secret"
			password := "custom-password-123"

			// 1. Create Secret
			secretData := map[string]any{
				"Name":     secretName,
				"Password": password,
			}
			secretManifest := renderTemplate(secretRconTemplate, secretData)
			kubectlSafeWithInput(secretManifest, "apply", "-f", "-")

			// 2. Create Minecraft CR
			mcData := map[string]any{
				"Name":                   name,
				"RconPasswordSecretName": secretName,
			}
			manifest := renderTemplate(minecraftRconTemplate, mcData)
			kubectlSafeWithInput(manifest, "apply", "-f", "-")

			// 2. Wait for Pod Ready
			waitStatefullSet("default", "mcing-"+name, 1)

			// 3. Ensure default secret is NOT created
			_, _, err := kubectl("get", "secret", "mcing-"+name+"-rcon-password")
			Expect(err).To(HaveOccurred()) // Should fail to find

			// 5. Verify RCON connection
			Eventually(func() error {
				_, stderr, err := kubectl(
					"exec",
					"mcing-"+name+"-0",
					"-c",
					"minecraft",
					"--",
					"rcon-cli",
					"--password",
					password,
					"list",
				)
				if err != nil {
					return fmt.Errorf("err: %w, stderr: %s", err, stderr)
				}
				return nil
			}).Should(Succeed())

			// Cleanup
			kubectlSafeWithInput(manifest, "delete", "-f", "-")
			kubectlSafeWithInput(secretManifest, "delete", "-f", "-")
		})
	})
}
