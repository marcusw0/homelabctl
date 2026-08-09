package cli

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/marcusw0/homelabctl/internal/config"
)

func parseCheck(args []string, opts GlobalOption) (Command, error) {
	if len(args) < 1 {
		return nil, errors.New("Expected command: dns|http|tcp|tls")
	}

	switch args[0] {
	case "dns":
		return parseDNSCheck(args[1:])
	case "http":
		return parseHTTPCheck(args[1:])
	case "tcp":
		return parseTCPCheck(args[1:])
	case "tls":
		return parseTLSCheck(args[1:])
	default:
		if len(args) != 1 {
			return nil, errors.New("check accepts exactly one server name")
		}

		serverName := args[0]

		var cfg config.Config
		if _, err := toml.DecodeFile(opts.ConfigPath, &cfg); err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}

		server, exists := cfg.Servers[serverName]
		if !exists {
			return nil, fmt.Errorf(
				"server %q not found in %s",
				serverName,
				opts.ConfigPath,
			)
		}
		if !server.Enabled {
			return nil, fmt.Errorf("server %q is disabled", serverName)
		}

		checkAll := CheckAllCmd{
			fqdn: server.FQDN,
			ip:   server.IP,
			port: server.Port,
		}
		return &checkAll, nil
	}

}
