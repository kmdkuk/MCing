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

func Whitelist(remoteConsole *rcon.RemoteConsole, action string, users []string) error {
	if action != "add" && action != "remove" {
		return fmt.Errorf("action must be add or remove. action: %s", action)
	}
	for _, user := range users {
		_, err := exec(remoteConsole, "whitelist", action, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func Whitelistlist(remoteConsole *rcon.RemoteConsole) ([]string, error) {
	// There are 2 whitelisted players: hoge, fuga
	liststr, err := exec(remoteConsole, "whitelist", "list")
	if err != nil {
		return nil, err
	}
	_, usertxt, ok := strings.Cut(liststr, ":")
	if !ok {
		return []string{}, nil
	}
	return strings.Split(strings.ReplaceAll(usertxt, " ", ""), ","), nil
}

func Op(remoteConsole *rcon.RemoteConsole, users []string) error {
	var errUsers []string
	for _, user := range users {
		out, err := exec(remoteConsole, "op", user)
		if err != nil {
			return err
		}
		if out == "That player does not exist" {
			errUsers = append(errUsers, user)
		}
	}
	if len(errUsers) > 0 {
		return fmt.Errorf("failed to add some users as operator users: %s", strings.Join(errUsers, ","))
	}
	return nil
}

func Deop(remoteConsole *rcon.RemoteConsole, users []string) error {
	for _, user := range users {
		_, err := exec(remoteConsole, "deop", user)
		if err != nil {
			return err
		}
	}
	return nil
}
