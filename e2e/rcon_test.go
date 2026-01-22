package e2e

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed testdata/minecraft-rcon.yaml.tmpl
var minecraftRconTemplate string

//go:embed testdata/secret-rcon.yaml.tmpl
var secretRconTemplate string

func renderTemplate(tmplStr string, data interface{}) []byte {
	tmpl, err := template.New("manifest").Parse(tmplStr)
	Expect(err).NotTo(HaveOccurred())
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	Expect(err).NotTo(HaveOccurred())
	return buf.Bytes()
}

func testRcon() {
	Describe("RCON Password", func() {
		It("should work with auto-generated secret", func() {
			name := "rcon-auto"
			// 1. Create Minecraft CR
			data := map[string]interface{}{
				"Name": name,
			}
			manifest := renderTemplate(minecraftRconTemplate, data)
			kubectlSafeWithInput(manifest, "apply", "-n", controllerNS, "-f", "-")

			// 2. Wait for Pod Ready
			waitStatefullSet(controllerNS, "mcing-"+name, 1)

			// 3. Get generated Secret
			out, _, err := kubectl("get", "secret", "mcing-"+name+"-rcon-password", "-n", controllerNS, "-o", "jsonpath={.data.rcon-password}")
			Expect(err).NotTo(HaveOccurred())
			passwordBytes, err := base64.StdEncoding.DecodeString(string(out))
			Expect(err).NotTo(HaveOccurred())
			password := string(passwordBytes)
			Expect(password).NotTo(BeEmpty())

			// 4. Verify RCON connection
			// rcon-cli needs some time to be ready after server start
			Eventually(func() error {
				_, stderr, err := kubectl("exec", "-n", controllerNS, "mcing-"+name+"-0", "-c", "minecraft", "--", "rcon-cli", "--password", password, "list")
				if err != nil {
					return fmt.Errorf("err: %v, stderr: %s", err, stderr)
				}
				return nil
			}, 3*time.Minute, 10*time.Second).Should(Succeed())
		})

		It("should work with user-specified secret", func() {
			name := "rcon-custom"
			secretName := "my-rcon-secret"
			password := "custom-password-123"

			// 1. Create Secret
			secretData := map[string]interface{}{
				"Name":     secretName,
				"Password": password,
			}
			secretManifest := renderTemplate(secretRconTemplate, secretData)
			kubectlSafeWithInput(secretManifest, "apply", "-n", controllerNS, "-f", "-")

			// 2. Create Minecraft CR
			mcData := map[string]interface{}{
				"Name":                   name,
				"RconPasswordSecretName": secretName,
			}
			manifest := renderTemplate(minecraftRconTemplate, mcData)
			kubectlSafeWithInput(manifest, "apply", "-n", controllerNS, "-f", "-")

			// 3. Wait for Pod Ready
			waitStatefullSet(controllerNS, "mcing-"+name, 1)

			// 4. Ensure default secret is NOT created
			_, _, err := kubectl("get", "secret", "mcing-"+name+"-rcon-password", "-n", controllerNS)
			Expect(err).To(HaveOccurred()) // Should fail to find

			// 5. Verify RCON connection
			Eventually(func() error {
				_, stderr, err := kubectl("exec", "-n", controllerNS, "mcing-"+name+"-0", "-c", "minecraft", "--", "rcon-cli", "--password", password, "list")
				if err != nil {
					return fmt.Errorf("err: %v, stderr: %s", err, stderr)
				}
				return nil
			}, 3*time.Minute, 10*time.Second).Should(Succeed())
		})
	})
}
