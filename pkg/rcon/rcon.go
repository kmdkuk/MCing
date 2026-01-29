package rcon

import (
	"fmt"
	"strings"

	"github.com/james4k/rcon"
)

// Console is an interface for RCON client.
type Console interface {
	Write(cmd string) (int, error)
	Read() (string, int, error)
}

// NewConn creates a new RCON connection.
func NewConn(hostPort, password string) (*rcon.RemoteConsole, error) {
	remoteConsole, err := rcon.Dial(hostPort, password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rcon server: %w", err)
	}
	return remoteConsole, nil
}

// edit from https://github.com/itzg/rcon-cli/blob/43ccb0311317dba9a99dd4836e4a274fbf993492/cli/entry.go#L98-L123

func exec(remoteConsole Console, command ...string) (string, error) {
	preparedCmd := strings.Join(command, " ")
	reqID, err := remoteConsole.Write(preparedCmd)
	if err != nil {
		return "", err
	}

	resp, respReqID, err := remoteConsole.Read()
	if err != nil {
		return "", fmt.Errorf("failed to read command: %w", err)
	}

	if reqID != respReqID {
		return "", fmt.Errorf("weird. this response is for another request. message: %s", resp)
	}

	return resp, nil
}

// Reload reloads the server.
func Reload(remoteConsole Console) error {
	_, err := exec(remoteConsole, "reload")
	if err != nil {
		return err
	}
	return nil
}

// WhitelistSwitch switches the whitelist on/off.
func WhitelistSwitch(remoteConsole Console, enabled bool) error {
	arg := "on"
	if !enabled {
		arg = "off"
	}
	_, err := exec(remoteConsole, "whitelist", arg)
	return err
}

// Whitelist adds or removes users from the whitelist.
func Whitelist(remoteConsole Console, action string, users []string) error {
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

// ListWhitelist lists whitelisted users.
func ListWhitelist(remoteConsole Console) ([]string, error) {
	// There are 2 whitelisted players: hoge, fuga
	liststr, err := exec(remoteConsole, "whitelist", "list")
	if err != nil {
		return nil, err
	}
	_, usertxt, ok := strings.Cut(liststr, ":")
	if !ok {
		return []string{}, nil
	}
	usertxt = strings.TrimSpace(usertxt)
	if usertxt == "" {
		return []string{}, nil
	}
	users := strings.Split(usertxt, ",")
	for i := range users {
		users[i] = strings.TrimSpace(users[i])
	}
	return users, nil
}

// Op adds users to the op list.
func Op(remoteConsole Console, users []string) error {
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

// Deop removes users from the op list.
func Deop(remoteConsole Console, users []string) error {
	for _, user := range users {
		_, err := exec(remoteConsole, "deop", user)
		if err != nil {
			return err
		}
	}
	return nil
}
