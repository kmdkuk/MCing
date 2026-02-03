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
	SetDefaultEventuallyPollingInterval(1 * time.Second)
	RunSpecs(t, "E2e Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	setupCluster()
	return nil
}, func(_ []byte) {})

var _ = Describe("mcing", func() {
	Context("opsWhitelist", testOpsWhitelist)
	Context("rcon", testRcon)
	Context("autopause", testAutoPause)
	Context("download", testDownload)
	Context("mcRouter", testMCRouter)
})
