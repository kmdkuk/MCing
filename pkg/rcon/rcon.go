package rcon

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/itzg/rcon-cli/cli"
)

func exec(out io.Writer, command ...string) {
	cli.Execute("localhost:25575", "minecraft", out, command...)
}

func Reload() {
	exec(os.Stdout, "reload")
}

func WhitelistSwitch(enabled bool) {
	arg := "on"
	if !enabled {
		arg = "off"
	}
	exec(os.Stdout, "whitelist", arg)
}

func Whitelistlist() []string {
	// There are 2 whitelisted players: hoge, fuga
	var b *bytes.Buffer
	exec(b, "whitelist", "list")
	return strings.Split(b.String(), ",")
}
