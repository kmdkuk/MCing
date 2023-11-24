package rcon

import (
	"os"

	"github.com/itzg/rcon-cli/cli"
)

func Reload() {
	cli.Execute("localhost:25575", "minecraft", os.Stdout, "reload")
}
