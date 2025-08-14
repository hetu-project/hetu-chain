<!--
parent:
  order: false
-->

<div align="center">
  <h1> Hetu Chain </h1>
</div>

Hetu Chain is a scalable, high-throughput blockchain that is fully compatible and interoperable with Ethereum.
It's built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/) with EVM compatibility, supporting both Cosmos and Ethereum ecosystems.

## Documentation

Our documentation is hosted in a [separate repository](https://github.com/hetu-project/docs).
Head over there and check it out.

## Prerequisites

- [Go 1.20+](https://golang.org/dl/)
- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [gcc](https://gcc.gnu.org/) (for cgo support)

## Installation

Follow these steps to install Hetu Chain from source:

### Clone the Repository

```bash
git clone https://github.com/hetu-project/hetu-chain.git
cd hetu-chain
```

### Build and Install

```bash
# Build the binary
make build

# Install the binary to your GOPATH
make install
```

After installation, verify that the binary is correctly installed:

```bash
hetud version
```

### Using Pre-built Binaries

Alternatively, you can download pre-built binaries from the latest [release](https://github.com/hetu-project/hetu/releases).

```bash
# Download the binary for your platform
wget https://github.com/hetu-project/hetu/releases/download/v0.x.x/hetud-v0.x.x-linux-amd64.tar.gz

# Extract the binary
tar -xzf hetud-v0.x.x-linux-amd64.tar.gz

# Move the binary to your PATH
sudo mv hetud /usr/local/bin/
```

## Deployment

### Local Deployment

To quickly set up a local development environment, use the `init.sh` script:

```bash
# Initialize a single-node local network
./init.sh

# Start the node
./start.sh
```

This script will:
1. Initialize the genesis file
2. Create a validator account
3. Add genesis accounts with test tokens
4. Configure the node
5. Start the node in development mode

### Remote Deployment

For multi-node deployments in a production environment:

1. **Initialize Validators**: Use the `init_validators.sh` script to set up validators in your network.

   ```bash
   ./init_validators.sh <remote_ip1> <remote_ip2> <remote_ip3> <remote_ip4>
   ```

2. **Start Archive Node**: For full historical data, start a node in archive mode.

   ```bash
   ./start_node_archive.sh
   ```

3. **Configure Networking**: Ensure proper firewall settings to allow P2P communication between nodes.

## Key Features

- **EVM Compatibility**: Full support for Ethereum smart contracts and tools
- **Cosmos SDK Integration**: Leverage Cosmos ecosystem features like IBC
- **Bittensor-inspired Consensus**: Advanced consensus mechanism for AI networks
- **Subnet Architecture**: Specialized subnets for different AI services
- **Alpha Token System**: Subnet-specific tokens for incentive alignment

## Community

The following chat channels and forums are a great spot to ask questions about Hetu Chain:

- [Open an Issue](https://github.com/hetu-project/hetu/issues)
- [Hetu Protocol](https://github.com/hetu-project#hetu-key-research)
- [Follow us on Twitter](https://x.com/hetu_protocol)

## Contributing

We welcome all contributions! There are many ways to contribute to the project, including but not limited to:

- Cloning code repo and opening a [PR](https://github.com/hetu-project/hetu/pulls).
- Submitting feature requests or [bugs](https://github.com/hetu-project/hetu/issues).
- Improving our product or contribution [documentation](https://github.com/hetu-project/hetu-docs).

For additional instructions, standards and style guides, please refer to the [Contributing](https://github.com/hetu-project/hetu-docs) document.
