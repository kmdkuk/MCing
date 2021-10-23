package e2e

import "os"

var (
	kubectlCmd = os.Getenv("KUBECTL")
)
