package cmd

import (
	"os"
	"path/filepath"

	"github.com/kmdkuk/mcing/pkg/constants"
)

func subMain() error {
	serverPropsPath := filepath.Join("/", "config", constants.ServerPropsName)
	os.Remove(constants.ServerPropsPath)
	return os.Symlink(serverPropsPath, constants.ServerPropsPath)
}
