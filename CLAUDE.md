# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**WatchCow** is a fnOS App Generator for Docker that monitors running Docker containers and automatically converts them into official fnOS applications using the `appcenter-cli install-local` command.

The system works by:
1. Monitoring Docker events (container start/stop)
2. Detecting containers with `watchcow.enable=true` label
3. Generating fnOS-compliant application directory structure
4. Installing the app via `appcenter-cli install-local`

## Build & Development Commands

### Building
```bash
# Build for current platform
make build

# Cross-compile for fnOS (Linux amd64)
make build-fnos

# Clean build artifacts
make clean
```

### Installation on fnOS
```bash
# Build for fnOS
make build-fnos

# Copy to fnOS host and run install script
scp build/watchcow-linux-amd64 install.sh watchcow.service user@fnos:/tmp/
ssh user@fnos "cd /tmp && sudo ./install.sh"
```

### Running
```bash
# Run directly (development)
make run

# Run as systemd service (on fnOS)
sudo systemctl start watchcow
sudo systemctl status watchcow

# View logs
journalctl -u watchcow -f

# Enable debug mode
watchcow --debug
```

### Testing
There are no automated tests. Testing is done manually:

```bash
# Create a test container with watchcow labels
docker run -d \
  --name test-nginx \
  --label watchcow.enable=true \
  --label watchcow.title="Test Nginx" \
  --label watchcow.port=80 \
  -p 8080:80 \
  nginx:alpine

# Verify watchcow detected and installed the app
journalctl -u watchcow | grep nginx
appcenter-cli list | grep nginx
```

## Architecture

### Core Components

**1. Docker Monitor (`internal/docker/monitor.go`)**
- Listens to Docker daemon events (start/stop/die/destroy)
- Detects containers with `watchcow.enable=true` label
- Triggers fpkgen to generate and install apps
- Tracks installed containers to avoid duplicates

**2. FPK Generator (`internal/fpkgen/`)**
- `types.go` - Core type definitions (AppConfig, VolumeMapping, etc.)
- `generator.go` - Main generator, extracts config from running containers
- `manifest.go` - Generates fnOS manifest file
- `compose.go` - Generates docker-compose.yaml from container config
- `scripts.go` - Generates cmd/ lifecycle scripts (start/stop/status)
- `config.go` - Generates config/privilege and config/resource files
- `ui.go` - Generates UI configuration (app/ui/config)
- `icons.go` - Downloads/generates app icons
- `installer.go` - Wraps appcenter-cli commands

### Generated App Structure

```
/tmp/watchcow-apps/<app-name>/
├── manifest                    # App metadata
├── app/
│   ├── docker-compose.yaml    # Container configuration
│   └── ui/
│       ├── config             # UI entry configuration
│       └── images/
│           └── ICON.PNG       # App icon
├── cmd/
│   └── main                   # Lifecycle script
├── config/
│   ├── privilege              # Run-as configuration
│   └── resource               # Docker project config
├── ICON.PNG                   # App icon (256x256)
└── ICON_256.PNG               # App icon (256x256)
```

### Key Data Flow

```
Docker Event (container start)
        ↓
Docker Monitor (detect watchcow.enable=true)
        ↓
FPK Generator (extract container config)
        ↓
Generate App Directory Structure
        ↓
appcenter-cli install-local
        ↓
App appears in fnOS App Center
```

### Smart Container Lifecycle

The generated `cmd/main` script handles the "container already running" scenario:
- On `start`: Checks if container exists and is running, skips if true
- On `stop`: Logs message but doesn't stop (container managed externally)
- On `status`: Returns proper exit codes for fnOS

### WatchCow Labels

Containers can be configured with labels:
- `watchcow.enable`: "true" to enable app generation (required)
- `watchcow.install`: "true" to auto-install via appcenter-cli
- `watchcow.title`: Display name in fnOS
- `watchcow.port`: Web UI port
- `watchcow.protocol`: "http" or "https" (default: http)
- `watchcow.path`: URL path (default: /)
- `watchcow.icon`: URL to download icon from
- `watchcow.category`: App category (default: "工具")
- `watchcow.description`: App description

## Platform Requirements

- **fnOS host** (Debian 12 based)
- **Docker** installed and running
- **appcenter-cli** available (fnOS built-in)
- Must run with access to Docker socket

## Development Guidelines

### When Adding Features
- Container config extraction is in `fpkgen/generator.go:GenerateFromContainer()`
- Manifest generation is in `fpkgen/manifest.go`
- Docker-compose generation is in `fpkgen/compose.go`
- Script generation is in `fpkgen/scripts.go`

### Debugging
- Use `--debug` flag for verbose logging
- Check generated app directory: `ls -la /tmp/watchcow-apps/<app-name>/`
- Verify manifest: `cat /tmp/watchcow-apps/<app-name>/manifest`
- Check appcenter-cli: `appcenter-cli list`

### Common Issues
- Container must have `watchcow.enable=true` label to be detected
- appcenter-cli must be available for auto-installation
- Icon download requires network access (falls back to placeholder)
- Container name becomes app name (sanitized to alphanumeric)

## Important Files

- `cmd/watchcow/main.go` - Entry point, flag parsing, initialization
- `internal/docker/monitor.go` - Docker event monitoring
- `internal/fpkgen/generator.go` - Main app generation logic
- `internal/fpkgen/installer.go` - appcenter-cli integration
- `Makefile` - Build commands
- `install.sh` - Installation script for fnOS
- `watchcow.service` - systemd service file
