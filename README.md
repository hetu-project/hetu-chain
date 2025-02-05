<!--
parent:
  order: false
-->

<div align="center">
  <h1> Hetu Chain </h1>
</div>

Hetu Chain is a scalable, high-throughput blockchain that is fully compatible and interoperable with Ethereum.
It's built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/) and implements
[Narwhal](https://github.com/MystenLabs/narwhal) + [Bullshark](https://github.com/MystenLabs/sui/tree/main/consensus/bullshark) consensus mechanism
for improved performance and scalability.

## Documentation

Our documentation is hosted in a [separate repository](https://github.com/hetu-project/docs) and can be found at [docs.hetu.org](https://docs.hetu.org).
Head over there and check it out.

**Note**: Requires [Go 1.20+](https://golang.org/dl/)

## Installation

For prerequisites and detailed build instructions
please read the [Installation](https://docs.hetu.org/protocol/hetu-cli) instructions.
Once the dependencies are installed, run:


make install

Or check out the latest [release](https://github.com/hetu-project/hetu/releases).

## Quick Start

To learn how the Hetu works from a high-level perspective,
go to the [Protocol Overview](https://docs.hetu.org/protocol) section from the documentation.
You can also check the instructions to [Run a Node](https://docs.hetu.org/protocol/hetu-cli#run-a-hetu-node).

## Deployment

### Local Deployment

To deploy locally, use the `local_node.sh` script. This script will set up a local environment for running the Hetu Chain node.

```bash
./local_node.sh
```

### Remote Deployment

For remote deployment, ensure that `hhubd` is available in the `PATH` on each machine. It is also recommended to set up SSH keys on the remote machines for secure and passwordless access.

1. `init_validators.sh`: This script initializes the validators required for the Hetu Chain. You need to provide the remote IPs for the 4 validators in the network as parameters.

   ```bash
   ./init_validators.sh <remote_ip1> <remote_ip2> <remote_ip3> <remote_ip4>
   ```

2. `start_node_archive.sh`: This script starts the node in archive mode.

   ```bash
   ./start_node_archive.sh
   ```

These scripts will help you set up and run the Hetu Chain on a remote server.

## Community

The following chat channels and forums are a great spot to ask questions about Hetu Chain:

- [Open an Issue](https://github.com/hetu-project/hetu/issues)
- [Hetu Protocol](https://github.com/hetu-project#hetu-key-research)
- [Follow us on Twitter](https://x.com/hetu_protocol)

## Contributing

We welcome all contributions! There are many ways to contribute to the project, including but not limited to:

- Cloning code repo and opening a [PR](https://github.com/hetu-project/hetu/pulls).
- Submitting feature requests or [bugs](https://github.com/hetu-project/hetu/issues).
- Improving our product or contribution [documentation](./docs/).
- Contributing [use cases](./demos/) to a feature request.

For additional instructions, standards and style guides, please refer to the [Contributing](./CONTRIBUTING.md) document.
