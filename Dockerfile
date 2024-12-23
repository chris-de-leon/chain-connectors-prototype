FROM docker.io/debian:12.8-slim

COPY /cc /usr/local/bin/

RUN apt-get update -y \
  && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends ca-certificates \
  && update-ca-certificates \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/* \
  && chmod +x /usr/local/bin/cc

ENTRYPOINT [ "/usr/local/bin/cc" ]
