package e2e

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	controllerNS = "mcing-system"
)

var (
	binDir = os.Getenv("BIN_DIR")
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(3 * time.Minute)
	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)
	RunSpecs(t, "E2e Suite")
}

var _ = Describe("mcing", func() {
	Context("bootstrap", testBootstrap)
	Context("opsWhitelist", testOpsWhitelist)
})
