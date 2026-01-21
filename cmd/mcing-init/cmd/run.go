package cmd

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kmdkuk/mcing/pkg/constants"
)

func subMain(cfg Config) error {
	if err := copyFiles(cfg); err != nil {
		return err
	}

	if err := buildSaveLazymcConfig(cfg); err != nil {
		return err
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
	return os.WriteFile(to, b, 0o600)
}

func copyFileWithExist(from, to string) error {
	if isFileExists(from) {
		return copyFile(from, to)
	}
	return nil
}

func copyExecutable(from, to string) error {
	b, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, b, 0o700) //nolint:gosec // 0o700 is required to write executable files.
}

func copyFiles(cfg Config) error {
	serverPropsPath := filepath.Join(constants.ConfigPath, constants.ServerPropsName)
	if err := os.Remove(constants.ServerPropsPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := copyFile(serverPropsPath, constants.ServerPropsPath); err != nil {
		return err
	}

	fileList := []struct {
		from  string
		to    string
		needs bool
		isBin bool
	}{
		{
			needs: true,
			isBin: false,
			from:  filepath.Join(constants.ConfigPath, constants.BanIPName),
			to:    constants.BanIPPath,
		},
		{
			needs: true,
			isBin: false,
			from:  filepath.Join(constants.ConfigPath, constants.BanPlayerName),
			to:    constants.BanIPPath,
		},
		{
			needs: true,
			isBin: false,
			from:  filepath.Join(constants.ConfigPath, constants.OpsName),
			to:    constants.OpsPath,
		},
		{
			needs: true,
			isBin: false,
			from:  filepath.Join(constants.ConfigPath, constants.WhiteListName),
			to:    constants.WhiteListPath,
		},
		{
			needs: true,
			isBin: false,
			from:  filepath.Join(constants.ConfigPath, constants.WhiteListName),
			to:    constants.WhiteListPath,
		},
		{
			needs: cfg.EnableLazyMC,
			isBin: true,
			from:  filepath.Join("/", constants.LazymcBinName),
			to:    constants.LazymcBinPath,
		},
	}

	for _, v := range fileList {
		if !v.needs {
			continue
		}
		if v.isBin {
			if err := copyExecutable(v.from, v.to); err != nil {
				return err
			}
			continue
		}
		if err := copyFileWithExist(v.from, v.to); err != nil {
			return err
		}
	}
	return nil
}

func buildSaveLazymcConfig(cfg Config) (err error) {
	if !cfg.EnableLazyMC {
		return nil
	}
	lazymcConfigPath := filepath.Join(constants.ConfigPath, constants.LazymcConfigName)
	if isFileExists(lazymcConfigPath) {
		if err = copyFile(lazymcConfigPath, constants.LazymcConfigPath); err != nil {
			return err
		}
	}

	rconPassword := os.Getenv("RCON_PASSWORD")
	if rconPassword == "" {
		return errors.New("required RCON_PASSWORD")
	}

	re := regexp.MustCompile("^password = .*")
	replacement := "password = \"" + rconPassword + "\""

	file, err := os.Open(constants.LazymcConfigPath)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil && !errors.Is(closeErr, os.ErrClosed) && err == nil {
			err = closeErr
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if re.MatchString(line) {
			line = replacement
		}
		lines = append(lines, line)
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	if err = file.Close(); err != nil {
		return err
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(constants.LazymcConfigPath, []byte(output), 0o600)
	if err != nil {
		return err
	}

	return nil
}
