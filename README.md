# Chain Connectors Prototype

## Archive Notice

I have decided to pause development on this project until further notice as I no longer have time to maintain it and add new features.

## Intro

The chain connectors project is a suite of web3 tooling that offers a standardized way to access real-time block data from a wide range of blockchains. The project currently supports the following chain families:

- All EVM-compatible chains
- All substrate-based chains
- Several non-EVM chains (e.g. Flow, Solana, etc.)

Each chain family has its own plugin which can be run using the chain connectors CLI tool. Under the hood, a chain family's plugin will:

1. Start a background process that subscribes to new blocks on the chain via a websocket connection
1. Start a gRPC server that allows clients to subscribe to the block data and interact with it in real-time

## Usage

Below we showcase several different ways that you can use the chain connectors CLI:

### Golang

List all plugins:

```go
go run github.com/chris-de-leon/chain-connectors-prototype/src/cli/apps/cli plugins list github
```

Run a plugin:

```go
go run github.com/chris-de-leon/chain-connectors-prototype/src/cli/apps/cli plugins run from-cli \
  --chain-wss "access.devnet.nodes.onflow.org:9000" \
  --plugin-id "flow"
```

### Docker

List all plugins:

```sh
docker run caffeineaddict333/chain-connectors-prototype:1.1.0 plugins list github
```

Run a plugin:

```sh
docker run caffeineaddict333/chain-connectors-prototype:1.1.0 plugins run from-cli \
  --chain-wss "access.devnet.nodes.onflow.org:9000" \
  --plugin-id "flow"
```

## Development

Enter a Nix shell with all necessary dev tools available:

```sh
make setup
```

Install dependencies:

```sh
make install
```

Upgrade dependencies:

```sh
make upgrade
```

Building the code:

```sh
make build
```

Building the code, artifacts, and local docker images:

```sh
make release.local
```

Testing:

```sh
# Test all projects with caching enabled
make test

# Test all projects even if there are no changes to the code
make test.no-cache
```

Invoking the CLI:

```sh
# You can invoke the CLI directly with Go:
go run ./src/cli/apps/cli/main.go version

# Or you can use some of the shortcuts in the Makefile:
make cli.plugins.run.from-config CHAIN=flow NETWORK=testnet

# Or you can invoke the CLI in a Docker container using `cc`:
make cli.docker
```

Regenerating Go files from protobuf spec:

```sh
make protogen
```
