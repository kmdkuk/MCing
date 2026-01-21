package e2e

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests

	"github.com/kmdkuk/mcing/pkg/constants"
)

//go:embed testdata/autopause.yaml.tmpl
var autopauseYAML string

//go:embed testdata/no-autopause.yaml.tmpl
var noAutopauseYAML string

func testAutoPause() {
	It("should auto-pause and auto-resume", func() {
		stsName := "mcing-autopause"
		manifest := renderTemplate(autopauseYAML, nil)
		kubectlSafeWithInput(manifest, "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)

		// Use port-forward to connect to the pod
		By("Connecting to trigger server start via port-forward")

		// Start port-forward in the background
		portForwardCmd, localPort, err := PortForwardCmd(
			context.Background(),
			"default",
			stsName+"-0",
			int(constants.ServerPort),
		)
		Expect(err).ShouldNot(HaveOccurred())
		portForwardCmd.Stdout = GinkgoWriter
		portForwardCmd.Stderr = GinkgoWriter
		err = portForwardCmd.Start()
		Expect(err).ShouldNot(HaveOccurred())
		defer func() {
			if portForwardCmd.Process != nil {
				_ = portForwardCmd.Process.Kill()
			}
		}()

		// Wait for port-forward to be ready
		By("Waiting for port-forward connection")
		Eventually(func(g Gomega) {
			//nolint:exhaustruct // only need timeout for test
			dialer := &net.Dialer{
				Timeout: 1 * time.Second,
			}
			conn, dialErr := dialer.DialContext(context.Background(), "tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
			g.Expect(dialErr).ShouldNot(HaveOccurred())
			err = conn.Close()
			g.Expect(err).ShouldNot(HaveOccurred())
		}).Should(Succeed())

		By("Verifying server starts on connection")
		Eventually(func(g Gomega) {
			// This sends a Minecraft login request which triggers lazymc to start the server
			err = mcTriggerServerStart("127.0.0.1", localPort)
			g.Expect(err).ShouldNot(HaveOccurred())
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			// Check logs for "Server is now online" (or check process if reliably working)
			// But wait, initially we want to verify start. Process check is fine here as it MUST start.
			stdout, _, execErr := kubectl("exec", stsName+"-0", "-c", "minecraft", "--", "pgrep", "java")
			g.Expect(execErr).ShouldNot(HaveOccurred()) // pgrep returns 0 if found
			g.Expect(string(stdout)).ShouldNot(BeEmpty())
		}).Should(Succeed())

		By("Waiting for auto-pause (timeout > 60s)")
		// Sleep 70s to exceed default lazymc 60s timeout + buffer
		time.Sleep(70 * time.Second)

		By("Verifying server is paused")
		Eventually(func(g Gomega) {
			stdout, _, execErr := kubectl("exec", stsName+"-0", "-c", "minecraft", "--", "pgrep", "java")
			g.Expect(execErr).Should(HaveOccurred())
			var exitErr *exec.ExitError
			g.Expect(errors.As(execErr, &exitErr)).Should(BeTrue())
			g.Expect(exitErr.ExitCode()).Should(Equal(1))
			g.Expect(string(stdout)).Should(BeEmpty())

			// https://github.com/timvisee/lazymc/blob/d058164aa6012b216eaae28e5581a6130dfeb7e6/src/server.rs#L524-L525
			// After stopping the server, there will be a short sleep before the status is updated, so wait for that.
			logs, _, execErr := kubectl("logs", stsName+"-0", "-c", "minecraft", "--tail=50")
			g.Expect(execErr).NotTo(HaveOccurred())
			g.Expect(logs).To(ContainSubstring("Server is now sleeping"), "lazymc has not finished cleanup yet")
		}).Should(Succeed())

		By("Connecting to resume server")
		// Connect again to trigger resume
		Eventually(func(g Gomega) {
			// This sends a Minecraft login request which triggers lazymc to start the server
			err = mcTriggerServerStart("127.0.0.1", localPort)
			g.Expect(err).ShouldNot(HaveOccurred())
		}).Should(Succeed())

		By("Verifying server is running")
		Eventually(func(g Gomega) {
			stdout, _, execErr := kubectl("exec", stsName+"-0", "-c", "minecraft", "--", "pgrep", "java")
			g.Expect(execErr).ShouldNot(HaveOccurred())
			g.Expect(string(stdout)).ShouldNot(BeEmpty())
		}).Should(Succeed())

		logs, _, err := kubectl("logs", stsName+"-0", "-c", "minecraft")
		Expect(err).NotTo(HaveOccurred())
		Expect(logs).To(ContainSubstring("> Failed to RCON server to sleep: authentication failed"))

		// Cleanup
		manifest = renderTemplate(autopauseYAML, nil)
		kubectlSafeWithInput(manifest, "delete", "-f", "-")
	})

	It("should run with disabled auto-pause", func() {
		stsName := "mcing-no-autopause"
		manifest := renderTemplate(noAutopauseYAML, nil)
		kubectlSafeWithInput(manifest, "apply", "-f", "-")
		waitStatefullSet("default", stsName, 1)

		By("Verifying server is running")
		Eventually(func(g Gomega) {
			stdout, _, err := kubectl("exec", stsName+"-0", "-c", "minecraft", "--", "pgrep", "java")
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(string(stdout)).ShouldNot(BeEmpty())
		}).Should(Succeed())

		// Cleanup
		kubectlSafeWithInput(manifest, "delete", "-f", "-")
	})
}
