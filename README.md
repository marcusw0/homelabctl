# Homelabctl

`homelabctl` is a CLI tool for managing and checking on homelab servers. You can do individual health checks for HTTP, TCP, TLS, and DNS or initialize a config file and start adding your servers in there. You can also keep track of runbooks/docs for each server and render the Markdown through `homelabctl` and even switch to editing them! I am still developing new features and will be adding a TUI dashboard soon.

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
- Interactive Markdown runbook viewing with editor integration
- Set custom width and styles with flags

## Installation

Install packages are available here
[latest GitHub release](https://github.com/marcusw0/homelabctl/releases/latest)

### Arch Linux

Download the `.pkg.tar.zst` file for your architecture, then run:

```bash
sudo pacman -U ./homelabctl-*.pkg.tar.zst
```

### Ubuntu and Debian

Download the `.deb` file for your architecture, then run:

```bash
sudo apt install ./homelabctl-*.deb
```

### Fedora and RHEL

Download the `.rpm` file for your architecture, then run:

```bash
sudo dnf install ./homelabctl-*.rpm
```

### Portable Linux and macOS installation

Download and extract the appropriate `.tar.gz` archive:

```bash
tar -xzf homelabctl_*.tar.gz
sudo install -m 0755 homelabctl /usr/local/bin/homelabctl
```

### Windows

Download the appropriate Windows `.zip` archive and extract it:

```powershell
Expand-Archive .\homelabctl_Windows_x86_64.zip -DestinationPath .\homelabctl
```

Move homelabctl.exe into a directory included in your PATH.

### Install with Go

```bash
go install github.com/marcusw0/homelabctl/cmd/homelabctl@latest
```

Ensure Go's binary directory is included in your PATH.

### Verify the installation

```bash
homelabctl --help
```

## Quick start

Initialize the configuration:

```bash
homelabctl config init
```

The default configuration location is:

- Linux: ~/.config/homelabctl/config.toml
- macOS: ~/Library/Application Support/homelabctl/config.toml
- Windows: %AppData%\homelabctl\config.toml

Try a health check:

```bash
homelabctl check tcp example.com
```

## Usage

All checks report health and protocol-specific information. Use `-v` or `--verbose` for additional details such as latency and check timestamps.

Every check will return the health. Protocol specific information such as HTTP status codes and TLS cert details are also displayed in their own checks and verbose mode will give back the most detail.

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

Check configured services:

```bash
homelabctl check service myserver
homelabctl check service --timeout 10s myserver
homelabctl check --all
```

Use global configuration and output options:

```bash
homelabctl --config ./homelabctl.toml list
homelabctl --verbose check service myNAS
```

### Config

Manage your toml config:

```bash
homelabctl config init
homelabctl config add <service>
homelabctl list
```

Example config:

```toml
[servers.MyServer]
fqdn = "myserver.example.com"
ip = "192.168.1.50"
port = 443
enabled = true
runbook = "runbooks/myserver.md"
```

### Runbooks

Services can reference a runbook file path:

```toml
[servers.MyServer]
fqdn = "myserver.example.com"
ip = "10.0.0.55"
port = 8443
enabled = true
runbook = "/path/to/runbook/myserver.md"
```

Runbook paths can either be absolute or resolved relative to the configuration directory.

View a configured service runbook:

```bash
homelabctl runbook myserver
homelabctl runbook --width 100 myserver
homelabctl runbook --style tokyo-night myserver
homelabctl runbook -w 80 -s ascii myserver
```

Controls:

- scroll: `j`/`k` or arrow keys
- edit: `e` - open runbook in `$EDITOR` (defaults to `vim`)
- exit: `q` or `esc`

After the editor exits, the markdown will be rendered again for viewing.
