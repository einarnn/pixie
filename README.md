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
| `--config <file>` | Path to the configuration YAML file (required) |
| `--config-stdin` | Read configuration YAML from stdin; does not update the source |
| `--debug` | Enable debug-level logging |
| `--info` | Enable info-level logging |
| `--insecure` | Skip TLS certificate verification |

## Building

```bash
go build -o pixiectl ./cmd/pixiectl/
```

## Commands

### List SGTs

Retrieve all SGTs from a device via the pxGrid TrustSec API.

```bash
pixiectl get-sgts --config config.yaml --device <device-name>
```

| Flag | Description |
|------|-------------|
| `--device` | Target device name (required) |

Output is pretty-printed JSON.

### Create SGT

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

### Delete SGT

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
