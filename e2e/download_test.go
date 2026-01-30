package e2e

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests
)

func testDownload() {
	ns := "default"

	DescribeTable("should download backup from a running server",
		func(name string, autoPause bool) {
			By("Creating a Minecraft server")
			manifest := renderTemplate(minecraftTmpl, map[string]any{
				"Name":      name,
				"Namespace": ns,
				"AutoPause": autoPause,
			})
			kubectlSafeWithInput(manifest, "apply", "-f", "-")

			By("Waiting for the server to be ready")
			waitStatefullSet(ns, "mcing-"+name, 1)

			By("Writing a dummy file to the server")
			dummyContent := "test-artifact"
			dummyPath := "/data/world/artifact.txt"
			kubectlSafe(
				"exec",
				"-n",
				ns,
				fmt.Sprintf("mcing-%s-0", name),
				"--",
				"sh",
				"-c",
				fmt.Sprintf("mkdir -p /data/world && echo %s > %s", dummyContent, dummyPath),
			)

			By("Downloading the backup")
			outputFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.tar.gz", name))
			defer os.Remove(outputFile)

			_, stderr := kubectlMcingSafe("download", name, "--namespace", ns, "--output", outputFile)
			Expect(string(stderr)).To(ContainSubstring("Download completed successfully."))

			By("Verifying the downloaded archive")
			verifyArchive(outputFile, "./world/artifact.txt", dummyContent)
		},
		Entry("with AutoPause disabled", "minecraft-autopause-disabled-running-java", false),
		Entry("with AutoPause enabled", "minecraft-autopause-enabled-sleeping-java", true),
	)
}
