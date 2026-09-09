# Fabric Network Orchestrator (FNO)

## Overview

Fabric Network Orchestrator (FNO) is a command-line tool and supporting domain-specific language (DSL) for specifying, validating, and deploying local Hyperledger Fabric networks.

FNO enables developers and researchers to describe Hyperledger Fabric topologies using a declarative specification rather than manually managing configuration files, certificates, deployment scripts, and chaincode lifecycle operations. Before deployment, specifications are validated against a set of formal validation constraints to detect configuration errors early.

The tool supports:

- Network specification through YAML, JSON, and TOML files
- Configuration validation
- Hyperledger Fabric artefact generation
- Certificate and MSP generation
- Network deployment and management
- Chaincode lifecycle management

FNO is primarily intended for:

- Research and experimentation
- Rapid prototyping
- Educational environments
- Local Hyperledger Fabric deployments

## Key Features

- Declarative network specification
- Validation of network configurations before deployment
- Support for channels, organisations, peers, orderers, and chaincodes
- Automated generation of Hyperledger Fabric artefacts
- Automated network provisioning
- Cross-platform execution
- No dependency on external Hyperledger Fabric CLI scripts

## Installation

### Using Prebuilt Binaries

Download the latest release from:

https://github.com/gca-research-group/fabric-network-orchestrator/releases

### Building from Source

Requirements:

- Docker
- Go 1.26 or later

```bash
git clone https://github.com/gca-research-group/fabric-network-orchestrator
cd fabric-network-orchestrator

go mod tidy
```

## Command Workflow

### Validate a Configuration

```bash
fno configuration validate --config samples/minimal-network.yml
```

### Generate Artefacts

```bash
fno artifacts generate --config samples/minimal-network.yml
```

Generates:

- Cryptographic material
- MSP structures
- TLS certificates
- Channel configuration artefacts
- Docker Compose descriptors

### Deploy a Network

```bash
fno network deploy --config samples/minimal-network.yml
```

This command performs:

1. Configuration validation
2. Identity generation
3. Artefact generation
4. Network provisioning
5. Channel creation and joining
6. Chaincode deployment (when configured)

### Manage Network Lifecycle

```bash
fno network up --config samples/minimal-network.yml
fno network down --config samples/minimal-network.yml
```

### Deploy Chaincodes

```bash
fno chaincode deploy --config samples/minimal-network.yml
```

### Clean Generated Artefacts

```bash
fno artifacts clean --config samples/minimal-network.yml
```

## Configuration Example

```yaml
output: output/minimal-network
network: minimal-network

capabilities:
  channel: V2_0
  orderer: V2_0
  application: V2_5

organizations:
  - name: Org1
    domain: org1.minimal-network.com
    bootstrap: true
    orderers:
      - name: Orderer
        subdomain: orderer
    peers:
      - name: Peer0
        subdomain: peer0
        isAnchor: true

  - name: Org2
    domain: org2.minimal-network.com
    peers:
      - name: Peer0
        subdomain: peer0
        isAnchor: true

profiles:
  - name: DefaultProfile
    organizations:
      - Org1
      - Org2

channels:
  - name: defaultchannel
    profile: DefaultProfile
```

## Validation Constraints

FNO validates specifications before deployment using the following 44 rules. The rule ID is included in each validation error and is also used by the experiment runner to identify the mutation applied to a scenario.

| Area | Rule ID | Constraint |
| --- | --- | --- |
| Output | `output.directory-name.invalid` | The output path must contain valid directory names. |
| Network | `network.name.required` | A network name is required. |
| Network | `network.name.invalid` | Network names must be compatible with generated Docker resources. |
| Organizations | `organizations.required` | At least one organization must be defined. |
| Organizations | `organization.name.required` | Every organization must have a name. |
| Organizations | `organization.domain.required` | Every organization must have a domain. |
| Organizations | `organization.domain.duplicate` | Organization domains must be unique. |
| Organizations | `organization.users.invalid` | Organization user counts cannot be negative. |
| Organizations | `domain.invalid` | Organization domains must be valid DNS domain names. |
| Organizations | `organization.name.duplicate` | Organization names must be unique. |
| Organizations | `organization.bootstrap.multiple` | At most one organization may be marked as bootstrap. |
| Certificate authorities | `certificate-authority.port.invalid` | Explicit CA exposed ports must be in the TCP range. |
| Peers | `peer.name.required` | Every peer must have a name. |
| Peers | `peer.subdomain.required` | Every peer must have a subdomain. |
| Peers | `peer.name.duplicate` | Peer names must be unique within an organization. |
| Peers | `peer.subdomain.duplicate` | Peer subdomains must be unique within an organization. |
| Peers | `peer.port.invalid` | Explicit peer exposed ports must be in the TCP range. |
| Peers | `peer.internal-port.invalid` | Explicit peer internal ports must be in the TCP range. |
| Peers | `peer.version.invalid` | A configured peer version must satisfy the channel capability's minimum binary version. |
| Orderers | `orderer.name.required` | Every orderer must have a name. |
| Orderers | `orderer.subdomain.required` | Every orderer must have a subdomain. |
| Orderers | `orderer.name.duplicate` | Orderer names must be unique within an organization. |
| Orderers | `orderer.port.invalid` | Explicit orderer exposed ports must be in the TCP range. |
| Orderers | `orderer.internal-port.invalid` | Explicit orderer internal ports must be in the TCP range. |
| Orderers | `orderer.version.invalid` | A configured orderer version must satisfy the channel capability's minimum binary version. |
| Orderers | `orderer.topology.required` | The topology must contain at least one orderer. |
| Chaincodes | `chaincode.name.required` | Every chaincode must have a name. |
| Chaincodes | `chaincode.path.required` | Every chaincode must have a source path. |
| Chaincodes | `chaincode.version.required` | Every chaincode must have a version. |
| Chaincodes | `chaincode.name.duplicate` | Chaincode names must be unique within a channel. |
| Profiles | `profile.name.required` | Every profile must have a name. |
| Profiles | `profile.name.duplicate` | Profile names must be unique. |
| Profiles | `profile.organizations.required` | Every profile must reference at least one organization. |
| Profiles | `profile.consensus-type.invalid` | Consensus type must be empty, `etcdraft`, or `BFT`. |
| Profiles | `profile.organization.undefined` | Every organization referenced by a profile must be defined. |
| Channels | `channel.name.required` | Every channel must have a name. |
| Channels | `channel.name.invalid` | Channel names must follow Hyperledger Fabric naming restrictions. |
| Channels | `channel.name.duplicate` | Channel names must be unique. |
| Channels | `channel.profile.required` | Every channel must reference a profile. |
| Channels | `channel.profile.undefined` | Every referenced channel profile must be defined. |
| Capabilities | `capability.channel.unsupported` | Channel capability must be `V2_0`, `V2_5`, or `V3_0`. |
| Capabilities | `capability.application.unsupported` | Application capability must be `V2_0`, `V2_5`, or `V3_0`. |
| Capabilities | `capability.orderer.unsupported` | Orderer capability must be `V2_0`, `V2_5`, or `V3_0`. |
| Networking | `exposed-port.conflict` | Positive exposed ports must be unique across CAs, peers, and orderers. |

Invalid configurations are rejected before artefacts are generated or infrastructure is provisioned.

## Experiment Runner

The developer-only experiment runner is implemented under [`internal/experiment`](./internal/experiment/) and launched from [`cmd/experiment-runner`](./cmd/experiment-runner/):

- `cli` wires the Cobra commands in small modules; the executable entry point only calls `cli.Execute()`.
- `generator` contains mutation operators for all 44 validation rules. Each operator starts from a valid seed configuration, introduces a targeted fault, and records the expected rule ID. Scenarios combine three compatible rule groups, with at most one mutation selected from each group; structurally incompatible mutations are explicitly excluded.
- `runner` streams the generated manifest, validates every scenario configuration, and checks that the reported validation rules include all mutations expected for that scenario.

Run its verification tests with:

```bash
go test ./cmd/experiment-runner ./internal/experiment/...
```

Create a complete seed, generate the scenario corpus, and validate it in separate calls:

```bash
go run ./cmd/experiment-runner seed generate --output seed.yaml
go run ./cmd/experiment-runner generate --seed seed.yaml --mutation-count 3 --output output
go run ./cmd/experiment-runner validate --output output
```

`seed generate` writes a bundled, deterministic YAML example with every configuration field explicitly represented, including optional settings. It defaults to `seed.yaml`, creates missing parent directories, and refuses to overwrite an existing file unless `--force` is supplied. Empty strings represent unused optional settings and empty lists represent absent topology entries. The seed includes three organizations and the three repository sample chaincodes; their paths are relative to the repository root. It generates configuration only, not chaincode sources or network artefacts.

For `generate`, the `--seed` flag is required and selects the valid YAML configuration to mutate. Existing seeds such as `samples/network-with-chaincode.yml` can also be used. Use `--mutation-count` to change the number of mutations per scenario (the default is `3`) and `--output` to select the generated corpus directory (the default is `output`). Generation writes mutated configurations to `<output-directory>/config/` and the scenario-to-rule mapping to `<output-directory>/scenarios.json`. It does not run verification or create or update `results.json`.

`validate` reads the existing manifest and scenario files from `--output` (default `output`) and writes `<output-directory>/results.json`. It needs neither the original seed nor a mutation count and does not regenerate scenarios. Each result is `passed` when all expected rules are reported, `partial` when only some are reported, or `failed` when none are reported or the configuration cannot be processed. Missing rules are included in the result. Validation exits with a non-zero status when any scenario is partial or failed. Generation retains its existing corpus overwrite behavior; validation replaces existing results.

Both phases process the corpus incrementally without retaining it in memory. Validation first counts manifest entries with a streaming read. Both `generate` and `validate` accept `--progress-interval N` to report progress every N scenarios (default `1000`; must be greater than zero). Start and completion are always reported.

Both commands also print UTC start and finish timestamps and elapsed duration, independently of the progress interval. Timing begins after argument validation, before reading the seed or counting the manifest, and ends after scenario/result files have been written. Elapsed duration uses Go's monotonic clock.

`<output>/metadata.json` records the current experiment using separate `generation` and `validation` sections. Each section contains:

- `parameters`: effective command flags, including defaults. Generation records `seed`, `mutationCount`, `output`, and `progressInterval`; validation records `output` and `progressInterval`.
- `workingDirectory`: the invocation's working directory, used to interpret relative paths.
- `startedAt` and `finishedAt`: UTC RFC3339 timestamps with fractional seconds when present.
- `durationSeconds`: elapsed duration in seconds.
- `status`: `running`, `succeeded`, or `failed`, plus `error` when the operation fails.

The command first saves a `running` record without `finishedAt` or `durationSeconds`, then updates it when processing returns. Ordinary failures, including partial or failed verification scenarios, receive a completed `failed` record. Metadata writes use a temporary file in the output directory and rename it after the complete JSON has been closed. Metadata-write errors cause a nonzero exit and preserve any operation error.

Starting generation replaces the metadata and clears previous validation metadata. Validation preserves generation metadata and replaces only its own section; an existing corpus without metadata receives a validation-only record. Malformed existing metadata is rejected during validation. One output directory represents the latest experiment, with no run history. Concurrent commands targeting the same output directory are unsupported, and interrupted processes may leave a `running` record. `seed generate` does not produce metadata.

Subcommands are required: replace the former `experiment-runner --seed ...` invocation with `experiment-runner generate --seed ...`, followed by `experiment-runner validate --output ...`. Use `--help` at the root, on `seed`, or on any individual command for usage.

## Samples

Network and chaincode samples are available in:

- [samples](./samples/)
- [chaincodes](./samples/chaincodes)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For any questions or issues, please open an issue on GitHub or contact the maintainers.
