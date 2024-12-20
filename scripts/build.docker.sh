#!/usr/bin/env bash

set -eo pipefail

if [ -z "$1" ]; then
  concurrency="0"
else
  concurrency="$1"
fi

cmds=()
for path in $(cd ./src/apps && find * -type f -name *.go); do
  src_dir="$(dirname ./src/apps/$path)"
  pth_dir="$(dirname $path)"
  cmds+=("docker build --build-arg APP_DIR=$src_dir --tag chain-connectors.${pth_dir//\//.}:$(uuidgen) .")
done

printf "\nBuilding ${#cmds[@]} apps...\n"
printf "%s\n" "${cmds[@]}" | xargs -P $concurrency -I {} bash -c {}
