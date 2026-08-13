# Homelabctl

`homelabctl` is currently a CLI tool for checking the health of your services. The long term goal is to build a TUI with service status, cert monitoring, and shortcuts to your documentation.

## Current Features

The CLI currently supports:

- HTTP status and latency checks
- TCP connectivity checks
- TLS certificate inspection
- DNS lookups
- creating a config.toml
- adding servers to config
- listing servers in config
- running health checks on servers in your config

## Usage

```bash
homelabctl check http --follow-redirects=false example.com
homelabctl check tcp 127.0.0.1:0
homelabctl check tls --timeout 2s example.com:443
homelabctl check dns example.com

homelabctl config init
homelabctl config add <server>
homelabctl list
homelabctl check service <server>
```

Every check will return the health, latency, and the time the check was performed. Protocol specific information such as HTTP status codes and TLS cert details are also displayed in their own checks.

## Project status

The network checks are working as expected. Configuration, inventories, and concurrency are in development. The TUI be added in a later stage.

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

### Phase 3: Configuration and CLI design — In progress

- [x] Load service definitions from TOML or JSON
- [x] Command-line flags
- [ ] Environment variables
- [ ] Built-in defaults
- [x] Add commands for listing and checking configured services
- [ ] Validate service names, targets, ports, durations, check types, and runbooks

Planned commands:

```bash
homelabctl list
homelabctl check service gitLab
homelabctl config add gitlab
homelabctl -v check service gitlab
```

### Phase 4: Concurrent checks

- Check multiple services concurrently
- Add homelabctl check --all
- Limit the number of simultaneous checks
- Cancel outstanding checks when Ctrl+C is pressed
- Verify concurrent behavior with Go's race detector

### Phase 5: Design and test hardening

- Refactor shared check behavior where useful
- Add dependency injection for network operations
- Expand timeout and cancellation tests
- Add configuration tests
- Experiment with fuzz testing target and configuration parsing

### Phase 6: Runbooks and terminal interface

- Render Markdown runbooks in the terminal
- Add runbook browsing, scrolling, search, and Vim-style navigation
- Build an interactive service dashboard

Planned commands:

```bash
homelabctl runbook openbao
homelabctl markdown ./runbooks/openbao.md
homelabctl dashboard
```

### Phase 7: Automation and releases

- Add GitLab CI for formatting, vetting, testing, and builds
- Run race detection in CI
- Scan dependencies with govulncheck
- Produce distributable Linux binaries
- Add a Hyprland shortcut for opening the dashboard
- Document installation and release procedures
