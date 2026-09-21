# Granular Configuration Validations Reference

This document outlines the complete set of **30 validation rules** implemented in `pkg/validate`. Each rule features a description and a **Universal Markdown / Unicode Math** representation, ensuring clean and perfect rendering in all standard markdown viewers without requiring external math extensions.

---

## Notation & Domains

Let:
* `C` be the configuration structure (`Config`).
* `Output(C)` be the configured output directory string.
* `Organizations` be the set of defined Organizations in `C`.
* `Peers(o)` be the set of Peers defined for organization `o ∈ Organizations`.
* `Orderers(o)` be the set of Orderers defined for organization `o ∈ Organizations`.
* `CA(o)` represent the Certificate Authority defined for organization `o ∈ Organizations`.
* `Chaincodes` be the set of defined Chaincodes in `C`.
* `Profiles` be the set of defined Profiles in `C`.
* `Channels` be the set of defined Channels in `C`.

For any entity `x`:
* `Name(x)` represents the name string.
* `Domain(o)` represents the domain string for an organization.
* `Subdomain(x)` represents the subdomain string for a peer or orderer.
* `ExposePort(x)` represents the exposed port integer.
* `Version(x)` represents the version string.
* `Path(x)` represents the source file path.
* `Bootstrap(o) ∈ {True, False}` indicates whether organization `o` is marked as the bootstrap organization.
* `ConsensusType(p)` represents the consensus type of profile `p`.
* `Organizations(p)` represents the set of organization names referenced by profile `p`.
* `ProfileRef(c)` represents the profile name referenced by channel `c`.

For Fabric Capabilities:
* `CapabilityMap = {"V2_0", "V2_5", "V3_0"}` is the set of supported capabilities.
* `ChannelCap(C)`, `ApplicationCap(C)`, `OrdererCap(C)` represent the capability strings configured.
* `MinBinaryVersion[cap]` represents the minimum binary version required for capability level `cap ∈ CapabilityMap`.

---

## 1. Rules (30 Total)

### Empty Output (`output.required`)
* **Description:** The target directory path for generating configuration artifacts must be specified.
* **Formal:**
  `Output(C) ≠ ""`

### Invalid Output Directory Name (`output.directory-name.invalid`)
* **Description:** Every component of the output directory path must be a valid directory name.
* **Formal:** Each non-navigation path component must match `^[a-zA-Z0-9._-]+$`.

### No Organizations (`organizations.required`)
* **Description:** At least one organization must be defined in the network topology.
* **Formal:**
  `|Organizations| ≥ 1`

### Empty Org Name (`organization.name.required`)
* **Description:** Every organization defined in the list must have a non-empty name.
* **Formal:**
  `∀ o ∈ Organizations, Name(o) ≠ ""`

### Empty Org Domain (`organization.domain.required`)
* **Description:** Every organization defined in the list must have a non-empty domain name.
* **Formal:**
  `∀ o ∈ Organizations, Domain(o) ≠ ""`

### Invalid CA Port (`certificate-authority.port.invalid`)
* **Description:** The Certificate Authority expose port for any organization must be greater than zero.
* **Formal:**
  `∀ o ∈ Organizations, ExposePort(CA(o)) > 0`

### Empty Peer Name (`peer.name.required`)
* **Description:** Every peer node defined in any organization must have a non-empty name.
* **Formal:**
  `∀ o ∈ Organizations, ∀ p ∈ Peers(o), Name(p) ≠ ""`

### Empty Peer Subdomain (`peer.subdomain.required`)
* **Description:** Every peer node defined in any organization must have a non-empty subdomain.
* **Formal:**
  `∀ o ∈ Organizations, ∀ p ∈ Peers(o), Subdomain(p) ≠ ""`

### Invalid Peer Port (`peer.port.invalid`)
* **Description:** The expose port for any peer node must be greater than zero.
* **Formal:**
  `∀ o ∈ Organizations, ∀ p ∈ Peers(o), ExposePort(p) > 0`

### Empty Orderer Name (`orderer.name.required`)
* **Description:** Every orderer node defined in any organization must have a non-empty name.
* **Formal:**
  `∀ o ∈ Organizations, ∀ ord ∈ Orderers(o), Name(ord) ≠ ""`

### Empty Orderer Subdomain (`orderer.subdomain.required`)
* **Description:** Every orderer node defined in any organization must have a non-empty subdomain.
* **Formal:**
  `∀ o ∈ Organizations, ∀ ord ∈ Orderers(o), Subdomain(ord) ≠ ""`

### Invalid Orderer Port (`orderer.port.invalid`)
* **Description:** The expose port for any orderer node must be greater than zero.
* **Formal:**
  `∀ o ∈ Organizations, ∀ ord ∈ Orderers(o), ExposePort(ord) > 0`

### Empty Chaincode Name (`chaincode.name.required`)
* **Description:** Every defined chaincode must have a non-empty name.
* **Formal:**
  `∀ cc ∈ Chaincodes, Name(cc) ≠ ""`

### Empty Chaincode Path (`chaincode.path.required`)
* **Description:** Every defined chaincode must have a non-empty path to its source code.
* **Formal:**
  `∀ cc ∈ Chaincodes, Path(cc) ≠ ""`

### Empty Chaincode Version (`chaincode.version.required`)
* **Description:** Every defined chaincode must specify a non-empty deployment version string.
* **Formal:**
  `∀ cc ∈ Chaincodes, Version(cc) ≠ ""`

### Empty Profile Orgs (`profile.organizations.required`)
* **Description:** Every configuration profile must include at least one organization reference in its list.
* **Formal:**
  `∀ p ∈ Profiles, |Organizations(p)| ≥ 1`

### Empty Channel Profile (`channel.profile.required`)
* **Description:** Every channel must reference a profile name.
* **Formal:**
  `∀ ch ∈ Channels, ProfileRef(ch) ≠ ""`

### Empty Channel Name (`channel.name.required`)
* **Description:** Every channel must have a non-empty name.
* **Formal:**
  `∀ ch ∈ Channels, Name(ch) ≠ ""`

### Unsupported Channel Capability (`capability.channel.unsupported`)
* **Description:** The channel capability string must exist in the supported capability registry.
* **Formal:**
  `ChannelCap(C) ∈ CapabilityMap`

### Unsupported Application Capability (`capability.application.unsupported`)
* **Description:** The application capability string must exist in the supported capability registry.
* **Formal:**
  `ApplicationCap(C) ∈ CapabilityMap`

### Unsupported Orderer Capability (`capability.orderer.unsupported`)
* **Description:** The orderer capability string must exist in the supported capability registry.
* **Formal:**
  `OrdererCap(C) ∈ CapabilityMap`

### Duplicate Org Name (`organization.name.duplicate`)
* **Description:** No two defined organizations in the network can have identical names.
* **Formal:**
  `∀ o₁, o₂ ∈ Organizations, o₁ ≠ o₂ ⟹ Name(o₁) ≠ Name(o₂)`

### Invalid Peer Version (`peer.version.invalid`)
* **Description:** If a peer version is specified, it must be greater than or equal to the stricter minimum required by the chosen channel and application capabilities.
* **Formal:**
  `∀ o ∈ Organizations, ∀ p ∈ Peers(o), Version(p) ≠ "" ⟹ Version(p) ≥ max(MinBinaryVersion[ChannelCap(C)], MinBinaryVersion[ApplicationCap(C)])`

### Invalid Orderer Version (`orderer.version.invalid`)
* **Description:** If an orderer version is specified, it must be greater than or equal to the stricter minimum required by the chosen channel and orderer capabilities.
* **Formal:**
  `∀ o ∈ Organizations, ∀ ord ∈ Orderers(o), Version(ord) ≠ "" ⟹ Version(ord) ≥ max(MinBinaryVersion[ChannelCap(C)], MinBinaryVersion[OrdererCap(C)])`

### No Orderer Topology (`orderer.topology.required`)
* **Description:** At least one organization in the configuration must define one or more orderer nodes.
* **Formal:**
  `∑ |Orderers(o)| ≥ 1  (for o ∈ Organizations)`

### Multiple Bootstrap Orgs (`organization.bootstrap.multiple`)
* **Description:** At most one organization can be marked as a bootstrap organization.
* **Formal:**
  `|{ o ∈ Organizations | Bootstrap(o) = True }| ≤ 1`

### Invalid Consensus Type (`profile.consensus-type.invalid`)
* **Description:** The profile's consensus type must either be left empty or match a supported type (`"etcdraft"` or `"BFT"`).
* **Formal:**
  `∀ p ∈ Profiles, ConsensusType(p) ∈ { "", "etcdraft", "BFT" }`

### Profile References Undefined Org (`profile.organization.undefined`)
* **Description:** Any organization name referenced inside a profile's organization list must match the name of a defined organization in the configuration.
* **Formal:**
  `∀ p ∈ Profiles, ∀ orgName ∈ Organizations(p), ∃ o ∈ Organizations s.t. Name(o) = orgName`

### Exposed Port Conflict (`exposed-port.conflict`)
* **Description:** No two services (Certificate Authority, Peer, or Orderer) can share the same host exposed port, preventing networking collisions at runtime.
* **Formal:**
  Let `E = { ExposePort(CA(o)) | o ∈ Organizations } ∪ { ExposePort(p) | o ∈ Organizations, p ∈ Peers(o) } ∪ { ExposePort(ord) | o ∈ Organizations, ord ∈ Orderers(o) }`
  `|E| = ∑_{o ∈ Organizations} (1 + |Peers(o)| + |Orderers(o)|)` (for expose port > 0)

### Invalid Channel Name (`channel.name.invalid`)
* **Description:** Channel names must conform to strict Hyperledger Fabric naming rules: lowercase alphanumeric characters, dots (`.`), or dashes (`-`), starting with a letter, and under 250 characters.
* **Formal:**
  `∀ ch ∈ Channels, MatchRegex("^[a-z][a-z0-9.-]{0,248}$", Name(ch))`
