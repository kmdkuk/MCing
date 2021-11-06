package cmd

import (
	"os"
	"path/filepath"

	"github.com/kmdkuk/mcing/pkg/constants"
)

func subMain() error {
	serverPropsPath := filepath.Join(constants.ConfigPath, constants.ServerPropsName)
	os.Remove(constants.ServerPropsPath)
	b, err := os.ReadFile(serverPropsPath)
	if err != nil {
		return err
	}
	return os.WriteFile(constants.ServerPropsPath, b, 0644)
}
