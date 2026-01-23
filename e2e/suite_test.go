package e2e

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests
)

const (
	controllerNS = "mcing-system"
)

//nolint:gochecknoglobals // test setup
var binDir = os.Getenv("BIN_DIR")

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(3 * time.Minute)
	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)
	RunSpecs(t, "E2e Suite")
}

var _ = Describe("mcing", Ordered, func() {
	Context("bootstrap", testBootstrap)
	Context("opsWhitelist", testOpsWhitelist)
	Context("rcon", testRcon)
})
