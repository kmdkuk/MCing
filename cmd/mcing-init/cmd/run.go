package cmd

import (
	"os"
	"path/filepath"

	"github.com/kmdkuk/mcing/pkg/constants"
)

func subMain() error {
	serverPropsPath := filepath.Join(constants.ConfigPath, constants.ServerPropsName)
	os.Remove(constants.ServerPropsPath)
	if err := copyFile(serverPropsPath, constants.ServerPropsPath); err != nil {
		return err
	}

	banIPPath := filepath.Join(constants.ConfigPath, constants.BanIPName)
	if !isFileExists(banIPPath) {
		if err := copyFile(banIPPath, constants.BanIPPath); err != nil {
			return err
		}
	}

	banPlayerPath := filepath.Join(constants.ConfigPath, constants.BanPlayerName)
	if !isFileExists(banPlayerPath) {
		if err := copyFile(banPlayerPath, constants.BanPlayerPath); err != nil {
			return err
		}
	}

	opsPath := filepath.Join(constants.ConfigPath, constants.OpsName)
	if !isFileExists(opsPath) {
		if err := copyFile(opsPath, constants.OpsPath); err != nil {
			return err
		}
	}

	whiteListPath := filepath.Join(constants.ConfigPath, constants.WhiteListName)
	if !isFileExists(whiteListPath) {
		if err := copyFile(whiteListPath, constants.WhiteListPath); err != nil {
			return err
		}
	}
	return nil
}

func isFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func copyFile(from, to string) error {
	b, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, b, 0644)
}
