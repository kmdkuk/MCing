package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/kmdkuk/mcing/pkg/constants"
)

var defaultServerProps = map[string]string{
	"enable-jmx-monitoring":             "false",
	"level-seed":                        "",
	"enable-command-block":              "true",
	"gamemode":                          "survival",
	"enable-query":                      "false",
	"generator-settings":                "",
	"level-name":                        "world",
	"motd":                              "A Vanilla Minecraft Server powered by MCing",
	"texture-pack":                      "",
	"pvp":                               "true",
	"generate-structures":               "true",
	"difficulty":                        "easy",
	"network-compression-threshold":     "256",
	"max-tick-time":                     "60000",
	"require-resource-pack":             "false",
	"max-players":                       "20",
	"use-native-transport":              "true",
	"online-mode":                       "true",
	"enable-status":                     "true",
	"allow-flight":                      "false",
	"broadcast-rcon-to-ops":             "true",
	"view-distance":                     "10",
	"max-build-height":                  "256",
	"server-ip":                         "",
	"resource-pack-prompt":              "",
	"allow-nether":                      "true",
	"enable-rcon":                       "true",
	"sync-chunk-writes":                 "true",
	"op-permission-level":               "4",
	"server-name":                       "Dedicated Server",
	"prevent-proxy-connections":         "false",
	"resource-pack":                     "",
	"entity-broadcast-range-percentage": "100",
	"player-idle-timeout":               "0",
	"rcon.password":                     "minecraft",
	"force-gamemode":                    "false",
	"rate-limit":                        "0",
	"hardcore":                          "false",
	"white-list":                        "false",
	"broadcast-console-to-ops":          "true",
	"spawn-npcs":                        "true",
	"spawn-animals":                     "true",
	"snooper-enabled":                   "true",
	"function-permission-level":         "2",
	"level-type":                        "DEFAULT",
	"text-filtering-config":             "",
	"spawn-monsters":                    "true",
	"enforce-whitelist":                 "false",
	"resource-pack-sha1":                "",
	"spawn-protection":                  "16",
	"max-world-size":                    "29999984",
}

var constServerProps = map[string]string{
	"rcon.port":   strconv.Itoa(int(constants.RconPort)),
	"query.port":  strconv.Itoa(int(constants.ServerPort)),
	"server-port": strconv.Itoa(int(constants.ServerPort)),
}

func GenServerProps(userProps map[string]string) (string, error) {
	serverProps := mergeProps(defaultServerProps, userProps)
	serverProps = mergeProps(serverProps, constServerProps)

	keys := make([]string, 0, len(serverProps))
	for k := range serverProps {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := new(strings.Builder)
	for _, k := range keys {
		_, err := fmt.Fprintf(b, "%s=%s\n", k, serverProps[k])
		if err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

func mergeProps(props1, props2 map[string]string) map[string]string {
	props := make(map[string]string)

	for k, v := range props1 {
		props[k] = v
	}

	for k, v := range props2 {
		props[k] = v
	}

	return props
}

func ParseServerPropsFromPath(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseServerProps(f)
}

func ParseServerProps(r io.Reader) (map[string]string, error) {
	props := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "#") {
			continue
		}
		if l == "" {
			continue
		}
		split := strings.SplitN(l, "=", 2)
		if len(split) != 2 {
			continue
		}
		props[split[0]] = split[1]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return props, nil
}
