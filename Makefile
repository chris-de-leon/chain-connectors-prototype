SUBSTRATE_PLUGIN_ARCHIVE_NAME="substrate-plugin"
SOLANA_PLUGIN_ARCHIVE_NAME="solana-plugin"
FLOW_PLUGIN_ARCHIVE_NAME="flow-plugin"
ETH_PLUGIN_ARCHIVE_NAME="eth-plugin"
CLI_ARCHIVE_NAME="cli"

docker.compose.up:
	@docker compose up --build -d

docker.compose.down:
	@docker compose down --remove-orphans

test.no-cache: docker.compose.up
	@go test -count=1 -v ./src/plugins/libs/...

test: docker.compose.up
	@go test -v ./src/plugins/libs/...

protogen:
	@bash ./scripts/protogen.sh

install:
	@go get -v ./... && go mod tidy

upgrade:
	@go get -v -u ./... && go mod tidy

clean:
	@go clean -x -i -r -cache -modcache

build:
	@goreleaser build --snapshot --verbose --clean

tag:
	@git tag -f "$$(go run ./src/cli/apps/cli/main.go version)"

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

cli.plugins.install.all: build
	@go run ./src/cli/apps/cli/main.go clean -a -f
	@go run ./src/cli/apps/cli/main.go plugins install local \
	  --plugin-path="$$(jq -erc --arg chain "substrate" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)" \
	  --plugin-path="$$(jq -erc --arg chain "solana" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)" \
	  --plugin-path="$$(jq -erc --arg chain "flow" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)" \
	  --plugin-path="$$(jq -erc --arg chain "eth" --arg os "$$(go env GOOS)" --arg arch "$$(go env GOARCH)" '.[] | select(.path | contains($$chain + "-plugin_" + $$os + "_" + $$arch)) | .path' ./dist/artifacts.json)"

# make cli.plugins.run.from-config CHAIN=flow NETWORK=testnet
cli.plugins.run-from-config:
	@go run ./src/cli/apps/cli/main.go \
		plugins run from-config \
			--config ./config.$(NETWORK).json \
			--name $(CHAIN)

# make cli.plugins.run.from-cli CHAIN=flow WSS="access.devnet.nodes.onflow.org:9000"
cli.plugins.run.from-cli:
	@go run ./src/cli/apps/cli/main.go \
		plugins run from-cli \
			--plugin-id $(CHAIN) \
			--chain-wss $(WSS)

cli.docker.local.run: release.local
	@IMG="$$(jq -erc --arg arch "$$(go env GOARCH)" '.[] | select(.type | contains("Docker Image")) | select(.name | contains($$arch)) | .name' ./dist/artifacts.json)" && \
	  docker run --rm -it --entrypoint /bin/bash "$$IMG"

# docker run --rm --entrypoint /bin/bash "$DOCKERHUB_USERNAME/$(basename $(go list -m)):$(go run ./src/cli/apps/cli/main.go version --no-prefix)" -c "cc plugins install github --plugin-id flow && cc plugins run from-cli --plugin-id flow --chain-wss access.devnet.nodes.onflow.org:9000"
cli.docker.hub.run:
	@docker run --rm -it --entrypoint /bin/bash "$$DOCKERHUB_USERNAME/$$(basename $$(go list -m)):$$(go run ./src/cli/apps/cli/main.go version --no-prefix)"
