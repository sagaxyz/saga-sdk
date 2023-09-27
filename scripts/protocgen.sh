#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep "option go_package" $file &> /dev/null ; then
      buf generate --template proto/buf.gen.gogo.yaml $file
    fi
  done
done

cp -r github.com/sagaxyz/saga-sdk/* ./
rm -rf github.com
