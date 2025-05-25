# Faino

Transactional deployment tool for containerized applications.

> [!WARNING]
> This is beta quality software. Use at your own risk.

## Features

- **Transaction management**: Transactional deployments with rollback on failure
- **Rollback capabilities**: Easy rollback to previous deployments
- **Multi-server deployment**: Deploy applications to multiple servers simultaneously

## Installation

### Homebrew

```bash
brew install lex-unix/tap/faino
```

### Go

```bash
go install github.com/lex-unix/faino/cmd/faino
```

## Quick Start

> [!NOTE]
> Faino doesn't install Docker on servers for you. You have to install Docker yourself.

1. **Initialize a new project**:

    ```bash
    faino init
    ```

2. **Configure your application** by editing `faino.yaml`:

    ```yaml
    service: my-app
    servers:
        - 192.168.1.10
        - 192.168.1.11
    registry:
        username: your-username
        password: your-password
    ```

3. **Setup your servers**:

    ```bash
    faino setup
    ```

4. **Deploy your application**:
    ```bash
    faino deploy
    ```

## Transaction Management

Faino manages deployments in a transactional manner. This means that if a deployment step fails on one of the servers, Faino will abort pending steps on other servers and begin rollback phase.
This ensures consistency across all target servers.

## Configuration

Faino uses a `faino.yaml` configuration file. Here's a complete example:

```yaml
# Application configuration
service: my-web-app

# Target servers
servers:
    - 192.168.1.10
    - 192.168.1.11
    - 192.168.1.12

# SSH configuration
ssh:
    user: root
    port: 22

# Registry configuration
registry:
    server: docker.io
    username: ${REGISTRY_USERNAME}
    password: ${REGISTRY_PASSWORD}

# Build configuration
build:
    dockerfile: .
    args:
        NODE_ENV: production
        API_URL: ${API_URL}

# Proxy/Load balancer configuration
proxy:
    container: traefik
    image: traefik:v3.1
    args:
        api.dashboard: true
    labels:
        traefik.enable: true

# Environment variables
env:
    NODE_ENV: production
    DATABASE_URL: ${DATABASE_URL}

# Secrets (will be expanded from environment variables)
secrets:
    API_KEY: ${API_KEY}
    DB_PASSWORD: ${DB_PASSWORD}

# Debug mode
debug: false
```

## Commands

### Deployment History & Rollback

```bash
# Deploy application
faino deploy

# View deployment history
faino history

# Rollback to specific version
faino rollback VERSION
```

### Application Management

```bash
# Start application containers
faino app start

# Stop application containers
faino app stop

# Restart application containers
faino app restart

# Show application status
faino app show

# Execute command in application container
faino app exec --interactive --host 192.168.0.1 "/bin/bash"
```

### Proxy Management

```bash
# Start proxy/load balancer
faino proxy start

# Stop proxy
faino proxy stop

# Restart proxy
faino proxy restart

# Show proxy status
faino proxy show

# View proxy logs
faino proxy logs

# Execute command in proxy container
faino proxy exec "traefik version"
```

## Global Flags

- `--debug, -d`: Enable debug output
- `--host`: Target specific host for command execution
- `--force`: Force non-transactional execution

## Configuration Options

### Required Fields

- `service`: Name of your service/application
- `servers`: List of target servers
- `registry.username`: Registry username
- `registry.password`: Registry password

### Optional Fields

- `image`: Docker image name (defaults to service name if not specified)
- `ssh.user`: SSH user (default: "root")
- `ssh.port`: SSH port (default: 22)
- `registry.server`: Registry server (default: "docker.io")
- `build.dockerfile`: Dockerfile path (default: ".")
- `build.args`: Build arguments
- `proxy.container`: Proxy container name (default: "traefik")
- `proxy.image`: Proxy image (default: "traefik:v3.1")
- `env`: Environment variables passed to `docker run`
- `secrets`: Secrets passed to `docker build`
- `debug`: Enable debug mode (default: false)

## Build System

Faino uses Docker Buildx for multi-platform builds with the following defaults:

- Builder: `faino-hybrid`
- Platform: `linux/amd64,linux/arm64`
- Driver: `docker-container`

## Examples

### Deploying Node.js container

```yaml
service: api-server
image: api
servers:
    - api1.example.com
    - api2.example.com
build:
    dockerfile: ./Dockerfile
    args:
        NODE_ENV: production
env:
    PORT: 3000
    DATABASE_URL: ${DATABASE_URL}
registry:
    server: ghcr.io
    username: ${REGISTRY_USER}
    password: ${REGISTRY_PASS}
```
