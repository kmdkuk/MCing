package watcher

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/log"
	"github.com/kmdkuk/mcing/pkg/rcon"
)

// Config represents the configuration for the watcher.
type Config struct {
	DataPath   string
	ConfigPath string
}

// NewDefaultConfig returns a new default configuration.
func NewDefaultConfig() Config {
	return Config{
		DataPath:   constants.DataPath,
		ConfigPath: constants.ConfigPath,
	}
}

// Watch watches the RCON server.
//
//nolint:gocognit // complex logic
func Watch(ctx context.Context, conn rcon.Console, interval time.Duration, cfg Config) error {
	tick := time.NewTicker(interval)
	defer tick.Stop()

	preConfig := map[string][]byte{}

	var err error

	preConfig[constants.ServerPropsName], err = os.ReadFile(filepath.Join(cfg.ConfigPath, constants.ServerPropsName))
	if err != nil {
		return err
	}

	for {
		select {
		case <-tick.C:
		case <-ctx.Done():
			log.Debug("quit")
			return nil
		}

		reload := false
		for k, v := range preConfig {
			path := filepath.Join(cfg.ConfigPath, k)
			current, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			if cmp.Equal(string(current), string(v)) {
				continue
			}
			dataPath := filepath.Join(cfg.DataPath, k)
			err = os.Remove(dataPath)
			if err != nil && !os.IsNotExist(err) {
				continue
			}
			err = os.WriteFile(dataPath, current, 0o600)
			if err != nil {
				continue
			}
			preConfig[k] = current
			reload = true
		}

		if reload {
			if err := rcon.Reload(conn); err != nil {
				return err
			}
		}
	}
}
