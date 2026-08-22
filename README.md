# Homelabctl

`homelabctl` is currently a CLI tool for checking the health of your services. The long term goal is to build a TUI with service status, cert monitoring, and shortcuts to your documentation.

## Current Features

The CLI currently supports:

- HTTP health checks with expected status, redirect, and timeout controls
- TCP connectivity checks for IPv4, IPv6, and hostnames
- TLS certificate validation and expiration reporting
- DNS lookups
- TOML-based service inventories
- Creating, validating, adding, and listing configured services
- Combined HTTP, DNS, TCP, and TLS service checks
- Checking all enabled services concurrently
- Verbose output with latency, timestamps, and protocol-specific details
- Markdown runbooks associated with configured services
- Terminal rendering with selectable style and width

## Runbooks

Services can reference a runbook file path:

```toml
[servers.MyServer]
fqdn = "myserver.example.com"
ip = "192.168.1.50
port = 443
enabled = true
runbook = "runbooks/myserver.md"
```

Runbook paths can either be absolute or resolved relative to the configuration directory.

## Usage

Run individual network checks:

```bash
homelabctl check http example.com
homelabctl check http --expect-status 204 --follow-redirects=false example.com
homelabctl check tcp --timeout 3s 127.0.0.1:80
homelabctl check tcp example.com:443
homelabctl check tcp '[::1]:443'
homelabctl check tls --timeout 2s --port 443 example.com
homelabctl check dns example.com
```

Manage the service inventory:

```bash
homelabctl config init
homelabctl config add <service>
homelabctl list
```

Check configured services:

```bash
homelabctl check service myserver
homelabctl check service --timeout 10s myserver
homelabctl check --all
```

Use global configuration and output options:

```bash
homelabctl --config ./homelabctl.toml list
homelabctl --verbose check service gitlab
```

View a configured service runbook:

```bash
homelabctl runbook myserver
homelabctl runbook --width 100 myserver
homelabctl runbook --style tokyo-night myserver
homelabctl runbook -w 80 -s ascii myserver
```

All checks report health and protocol-specific information. Use `-v` or `--verbose` for additional details such as latency and check timestamps.

Every check will return the health. Protocol specific information such as HTTP status codes and TLS cert details are also displayed in their own checks and verbose mode will give back the most detail.

## Project status

The core CLI is functional. Homelabctl supports one-off network checks, validated TOML config, individual service checks, and concurrent checks across all enabled services.

Development is currently focused on design and test hardening, including additional configuration, timeout, cancellation, and network tests. Basic runbook rendering is now supported. Interactive browsing, scrolling, and search are in development as well as the dashboard.

## Roadmap

### Phase 1: Basic HTTP checks — Complete

- Run HTTP health checks from the command line
- Report status code, latency, and health
- Return useful errors for failed requests
- Add structured HTTP results
- Add tests

### Phase 2: Network checks — Complete

- HTTP status and latency checks
- TCP connectivity checks
- TLS certificate inspection
- DNS lookups
- Context cancellation and per-check timeouts
- Structured results with health, latency, and timestamps
- Certificate expiration reporting
- HTTP redirect reporting
- Automated tests for TCP, HTTP, and DNS checks

### Phase 3: Configuration and CLI design — Complete

- Load service definitions from TOML
- Command-line flags
- Built-in defaults
- Add commands for listing and checking configured services

Planned commands:

```bash
homelabctl list
homelabctl check service gitlab
homelabctl config add gitlab
homelabctl -v check service gitlab
```

### Phase 4: Concurrent checks — Complete

- Check multiple services concurrently
- Add homelabctl check --all
- Limit the number of simultaneous checks
- Cancel outstanding checks when Ctrl+C is pressed
- Verify concurrent behavior with Go's race detector
- Validate service hostnames, IP addresses, ports, and check types

### Phase 5: Design and test hardening  — Complete

- Refactor shared check behavior where useful
- Expand timeout and cancellation tests
- Add configuration tests
- Experiment with fuzz testing

### Phase 6: Runbooks and terminal interface — In progress

- [x] Render service runbooks in the terminal
- [x] Support selectable rendering styles and widths
- [ ] Add runbook browsing, scrolling, search, and Vim-style navigation
- [ ] Build an interactive service dashboard

Planned commands:

```bash
homelabctl runbook -w 80 myserver
homelabctl dashboard
```

### Phase 7: Automation and releases

- Scan dependencies with govulncheck
- Produce distributable Linux binaries
- Add a Hyprland shortcut for opening the dashboard
- Document installation and release procedures
