#!/usr/bin/env bash

set -eo pipefail

if [ -z "$1" ]; then
  concurrency="0"
else
  concurrency="$1"
fi

cmds=()
for path in $(cd ./src/apps && find * -type f -name *.go); do
  bin="./bin/apps/${path/main.go/bin}"
  src="./src/apps/$path"
  cmds+=("echo \"Building $src\" && go build -o $bin $src && echo \"Created $bin\"")
done

printf "\nBuilding ${#cmds[@]} apps...\n"
printf "%s\n" "${cmds[@]}" | xargs -P $concurrency -I {} bash -c {}
