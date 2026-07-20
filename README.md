# pixiectl

CLI tool for managing pxGrid Cloud devices and TrustSec Security Group Tags (SGTs).

## Running the container

Build the local-architecture image with Task:

```bash
task pixiectl-image
```

The container reads its YAML configuration from stdin. For example:

```bash
cat config.yaml | docker run --rm -i \
  pixiectl:latest \
  get-sgts \
  --config-stdin \
  --device my-ise-node
```

The `--config-stdin` option does not write the configuration back to disk. Use a config containing an already-linked tenant (`id`, `name`, and `token`); initial OTP linking requires the writable file-based `--config` option.

## Prerequisites

A YAML configuration file is required for all commands. It contains app credentials and tenant information:

```yaml
app:
  id: "<app-id>"
  apiKey: "<api-key>"
  globalFQDN: "<global-fqdn>"
  regionalFQDN: "<regional-fqdn>"
  readStream: "<read-stream>"
  writeStream: "<write-stream>"

tenant:
  otp: "<one-time-password>"   # only needed for first-time tenant linking
  id: "<tenant-id>"
  name: "<tenant-name>"
  token: "<tenant-token>"
```

On first use, supply the `otp` field to link the tenant. After linking, `pixiectl` updates the config file with the tenant `id`, `name`, and `token` for subsequent runs.

## Global Flags

| Flag | Description |
|------|-------------|
| `--config <file>` | Path to the configuration YAML file; mutually exclusive with `--config-stdin` |
| `--config-stdin` | Read configuration YAML from stdin; does not update the source |
| `--debug` | Enable debug-level logging |
| `--info` | Enable info-level logging |
| `--insecure` | Skip TLS certificate verification |

## Building

```bash
go build -o pixiectl ./cmd/pixiectl/
```

## Commands

All commands require configuration. Use either `--config <file>` or
`--config-stdin`, but not both. The examples below use a file; for the
container, pipe the YAML to stdin and use `--config-stdin`:

```bash
cat config.yaml | docker run --rm -i pixiectl:latest \
  list-devices --config-stdin
```

### `get-sgts`

Retrieve all SGTs from a device via the pxGrid TrustSec API.

```bash
pixiectl get-sgts --config config.yaml --device <device-name>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |

Output is pretty-printed JSON.

### `create-sgt`

Create a new SGT on a device via the ERS API.

```bash
pixiectl create-sgt --config config.yaml \
  --device <device-name> \
  --name <sgt-name> \
  --description <sgt-description> \
  --tag <sgt-value>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |
| `--name` | Name for the new SGT (required) |
| `--description` | Description for the new SGT (required) |
| `--tag` | Numeric SGT value, must be non-zero (required) |

### `delete-sgt`

Delete an SGT by name from a device. The command first looks up the SGT ID by name, then issues a delete.

```bash
pixiectl delete-sgt --config config.yaml \
  --device <device-name> \
  --name <sgt-name>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |
| `--name` | Name of the SGT to delete (required) |

### `list-devices`

Lists all devices registered for the configured tenant. The output includes
the device name, type, and status.

```bash
pixiectl list-devices --config config.yaml
```

This command has no command-specific flags.

### `get-sessions`

Lists established sessions for a device through the pxGrid session API. The
response is printed as pretty-printed JSON.

```bash
pixiectl get-sessions --config config.yaml --device <device-name>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |

### `get-anc-policies`

Lists ANC policies available through the pxGrid ANC API. The response is
printed as pretty-printed JSON.

```bash
pixiectl get-anc-policies --config config.yaml --device <device-name>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |

### `apply-anc-policy`

Applies an ANC policy to a client identified by its MAC address.

```bash
pixiectl apply-anc-policy --config config.yaml \
  --device <device-name> \
  --name <policy-name> \
  --mac <mac-address>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |
| `--name` | ANC policy name (required) |
| `--mac` | Client MAC address (required) |

### `clear-anc-policy`

Clears the ANC policy from a client identified by its MAC address.

```bash
pixiectl clear-anc-policy --config config.yaml \
  --device <device-name> \
  --mac <mac-address>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |
| `--mac` | Client MAC address (required) |

### `run`

Connects to pxGrid Cloud for a device and waits for messages or termination.
This is a long-running command; press `Ctrl-C` to stop it.

```bash
pixiectl run --config config.yaml --device <device-name>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |

### Built-in commands

Cobra also provides help and shell-completion commands:

```bash
pixiectl help
pixiectl help <command>
pixiectl completion bash
pixiectl completion zsh
pixiectl completion fish
pixiectl completion powershell
```

`completion` prints shell-completion code for the selected shell.

## Examples

```bash
# List all SGTs on a device
pixiectl get-sgts --config config.yaml --device my-ise-node --info

# Create a new SGT
pixiectl create-sgt --config config.yaml \
  --device my-ise-node \
  --name "Employees" \
  --description "Corporate employees" \
  --tag 100

# Delete an SGT
pixiectl delete-sgt --config config.yaml \
  --device my-ise-node \
  --name "Employees"
```
