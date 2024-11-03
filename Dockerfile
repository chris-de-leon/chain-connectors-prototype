FROM golang:1.23.2-bullseye AS build
ARG APP_DIR
WORKDIR /usr/chain-connectors
COPY ./src ./src
COPY ./go.mod ./go.sum ./
RUN --mount=type=cache,target="/go/pkg/mod" \
  go mod download -x && \
  go build -o /usr/chain-connectors/bin /usr/chain-connectors/${APP_DIR}/main.go

# The bullseye docker image doesn't come with ca-certificates by default :(
# We need to install it manually:
#
#  https://github.com/debuerreotype/docker-debian-artifacts/issues/15#issuecomment-634423712
#
FROM debian:11.9-slim
WORKDIR /workspace
COPY --from=build /usr/chain-connectors/bin /usr/chain-connectors/bin
RUN apt-get update -y \
  && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends ca-certificates \
  && update-ca-certificates \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/chain-connectors/bin"]
