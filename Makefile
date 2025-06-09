CLI_VERSION = $(shell cat ./VERSION | tr -d '\n')
SUBSTRATE_PLUGIN_ARCHIVE_NAME="substrate-plugin"
SOLANA_PLUGIN_ARCHIVE_NAME="solana-plugin"
FLOW_PLUGIN_ARCHIVE_NAME="flow-plugin"
ETH_PLUGIN_ARCHIVE_NAME="eth-plugin"
OUTPUTS_DIR = $(PWD)/bin
BUILDER_DIR = $(OUTPUTS_DIR)/builder
CLI_ARCHIVE_NAME="cli"
MAKEFLAGS += --no-print-directory
SHELL = /bin/bash -eo pipefail

.PHONY: docker.compose.up
docker.compose.up:
	@docker compose up --build -d

.PHONY: docker.compose.down
docker.compose.down:
	@docker compose down --remove-orphans

.PHONY: nixupdate
nixupdate:
	@nix flake update

.PHONY: nixlock
nixlock:
	@nix flake lock

.PHONY: nixfmt
nixfmt:
	@nix fmt .

.PHONY: test.no-cache
test.no-cache: docker.compose.up
	@go test -count=1 -v ./src/plugins/libs/...

.PHONY: test
test: docker.compose.up
	@go test -v ./src/plugins/libs/...

.PHONY: protogen
protogen:
	@bash ./scripts/protogen.sh

.PHONY: setup
setup:
	@nix develop

.PHONY: install
install:
	@go get -v ./... && go mod tidy

.PHONY: upgrade
upgrade:
	@go get -v -u ./... && go mod tidy

.PHONY: clean
clean:
	@go clean -x -i -r -cache -modcache
	@rm -rf ./dist
	@rm -rf ./bin

.PHONY: build
build:
	@goreleaser build --snapshot --verbose --clean

.PHONY: image
image:
	@mkdir -p $(BUILDER_DIR) && nix build ".#docker-sandbox" --print-out-paths --out-link $(BUILDER_DIR)/cc-sandbox.tar.gz

.PHONY: unload
unload:
	@docker image rm --force "cc:$(CLI_VERSION)-sandbox"

.PHONY: load
load:
	@docker load < $(BUILDER_DIR)/cc-sandbox.tar.gz

.PHONY: bin
bin:
	@mkdir -p $(BUILDER_DIR) && nix build ".#cc" --print-out-paths --out-link $(BUILDER_DIR)/cc

.PHONY: tag
tag:
	@git tag -f "$(CLI_VERSION)"

.PHONY: sandbox
sandbox: unload
sandbox: image
sandbox: load
sandbox:
	@docker run --rm -it \
		-v $(PWD)/config.mainnet.json:/workspace/config.mainnet.json:ro \
		-v $(PWD)/config.testnet.json:/workspace/config.testnet.json:ro \
		-v /etc/ssl/private:/etc/ssl/private:ro \
		-v /etc/ssl/certs:/etc/ssl/certs:ro \
		-w /workspace \
		"cc:$(CLI_VERSION)-sandbox" \
		bash

.PHONY: release.github
release.github: tag
	@\
		SUBSTRATE_PLUGIN_ARCHIVE_NAME="$(SUBSTRATE_PLUGIN_ARCHIVE_NAME)" \
		SOLANA_PLUGIN_ARCHIVE_NAME="$(SOLANA_PLUGIN_ARCHIVE_NAME)" \
		FLOW_PLUGIN_ARCHIVE_NAME="$(FLOW_PLUGIN_ARCHIVE_NAME)" \
		ETH_PLUGIN_ARCHIVE_NAME="$(ETH_PLUGIN_ARCHIVE_NAME)" \
		CLI_ARCHIVE_NAME="$(CLI_ARCHIVE_NAME)" \
		SKIP_GITHUB="false" \
		SKIP_DOCKER="true" \
		goreleaser release \
			--skip=validate,docker \
			--verbose \
			--clean

.PHONY: release.docker
release.docker: tag
	@\
		SUBSTRATE_PLUGIN_ARCHIVE_NAME="$(SUBSTRATE_PLUGIN_ARCHIVE_NAME)" \
		SOLANA_PLUGIN_ARCHIVE_NAME="$(SOLANA_PLUGIN_ARCHIVE_NAME)" \
		FLOW_PLUGIN_ARCHIVE_NAME="$(FLOW_PLUGIN_ARCHIVE_NAME)" \
		ETH_PLUGIN_ARCHIVE_NAME="$(ETH_PLUGIN_ARCHIVE_NAME)" \
		CLI_ARCHIVE_NAME="$(CLI_ARCHIVE_NAME)" \
		SKIP_GITHUB="true" \
		SKIP_DOCKER="false" \
		goreleaser release \
			--skip=validate \
			--verbose \
			--clean

.PHONY: release.local
release.local:
	@\
		SUBSTRATE_PLUGIN_ARCHIVE_NAME="$(SUBSTRATE_PLUGIN_ARCHIVE_NAME)" \
		SOLANA_PLUGIN_ARCHIVE_NAME="$(SOLANA_PLUGIN_ARCHIVE_NAME)" \
		FLOW_PLUGIN_ARCHIVE_NAME="$(FLOW_PLUGIN_ARCHIVE_NAME)" \
		ETH_PLUGIN_ARCHIVE_NAME="$(ETH_PLUGIN_ARCHIVE_NAME)" \
		CLI_ARCHIVE_NAME="$(CLI_ARCHIVE_NAME)" \
		SKIP_GITHUB="true" \
		SKIP_DOCKER="true" \
		goreleaser release \
			--snapshot \
			--verbose \
			--clean

.PHONY: release.all.strict
release.all.strict: tag
	@\
		SUBSTRATE_PLUGIN_ARCHIVE_NAME="$(SUBSTRATE_PLUGIN_ARCHIVE_NAME)" \
		SOLANA_PLUGIN_ARCHIVE_NAME="$(SOLANA_PLUGIN_ARCHIVE_NAME)" \
		FLOW_PLUGIN_ARCHIVE_NAME="$(FLOW_PLUGIN_ARCHIVE_NAME)" \
		ETH_PLUGIN_ARCHIVE_NAME="$(ETH_PLUGIN_ARCHIVE_NAME)" \
		CLI_ARCHIVE_NAME="$(CLI_ARCHIVE_NAME)" \
		SKIP_GITHUB="false" \
		SKIP_DOCKER="false" \
		goreleaser release \
			--verbose \
			--clean

.PHONY: release.all
release.all: tag
	@\
		SUBSTRATE_PLUGIN_ARCHIVE_NAME="$(SUBSTRATE_PLUGIN_ARCHIVE_NAME)" \
		SOLANA_PLUGIN_ARCHIVE_NAME="$(SOLANA_PLUGIN_ARCHIVE_NAME)" \
		FLOW_PLUGIN_ARCHIVE_NAME="$(FLOW_PLUGIN_ARCHIVE_NAME)" \
		ETH_PLUGIN_ARCHIVE_NAME="$(ETH_PLUGIN_ARCHIVE_NAME)" \
		CLI_ARCHIVE_NAME="$(CLI_ARCHIVE_NAME)" \
		SKIP_GITHUB="false" \
		SKIP_DOCKER="false" \
		goreleaser release \
		  --skip=validate \
			--verbose \
			--clean


.PHONY: cli.clean.all
cli.clean.all:
	@go run ./src/cli/apps/cli/main.go clean -a -f

.PHONY: cli.plugins.install.all
cli.plugins.install.all: build
	@go run ./src/cli/apps/cli/main.go plugins install local \
	  --plugin-path="$$(jq -erc --arg chain "substrate" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)" \
	  --plugin-path="$$(jq -erc --arg chain "solana" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)" \
	  --plugin-path="$$(jq -erc --arg chain "flow" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)" \
	  --plugin-path="$$(jq -erc --arg chain "eth" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)" \
	  --clean

# make cli.plugins.run.from-config CHAIN=flow NETWORK=testnet
.PHONY: cli.plugins.run.from-config
cli.plugins.run.from-config:
	@go run ./src/cli/apps/cli/main.go \
		plugins run from-config \
			--config ./config.$(NETWORK).json \
			--name $(CHAIN)

# make cli.plugins.run.from-cli CHAIN=flow WSS="access.devnet.nodes.onflow.org:9000"
.PHONY: cli.plugins.run.from-cli
cli.plugins.run.from-cli:
	@go run ./src/cli/apps/cli/main.go \
		plugins run from-cli \
			--plugin-id $(CHAIN) \
			--chain-wss $(WSS)

