package rcon

import (
	"fmt"
	"strings"

	"github.com/james4k/rcon"
)

func NewConn(hostPort, password string) (*rcon.RemoteConsole, error) {
	remoteConsole, err := rcon.Dial(hostPort, password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rcon server: %w", err)
	}
	return remoteConsole, nil
}

// edit from https://github.com/itzg/rcon-cli/blob/43ccb0311317dba9a99dd4836e4a274fbf993492/cli/entry.go#L98-L123

func exec(remoteConsole *rcon.RemoteConsole, command ...string) (string, error) {
	preparedCmd := strings.Join(command, " ")
	reqId, err := remoteConsole.Write(preparedCmd)
	if err != nil {
		return "", err
	}

	resp, respReqId, err := remoteConsole.Read()
	if err != nil {
		return "", fmt.Errorf("failed to read command: %w", err)

	}

	if reqId != respReqId {
		return "", fmt.Errorf("weird. this response is for another request. message: %s", resp)
	}

	return resp, nil
}

func Reload(remoteConsole *rcon.RemoteConsole) error {
	str, err := exec(remoteConsole, "reload")
	if err != nil {
		return err
	}
	fmt.Println(str)
	return nil
}

func WhitelistSwitch(remoteConsole *rcon.RemoteConsole, enabled bool) error {
	arg := "on"
	if !enabled {
		arg = "off"
	}
	_, err := exec(remoteConsole, "whitelist", arg)
	return err
}

func Whitelistlist(remoteConsole *rcon.RemoteConsole) ([]string, error) {
	// There are 2 whitelisted players: hoge, fuga
	liststr, err := exec(remoteConsole, "whitelist", "list")
	if err != nil {
		return nil, err
	}
	return strings.Split(liststr, ","), nil
}
