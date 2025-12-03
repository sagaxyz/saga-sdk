#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/chainlet/types/expected_keepers.go -package testutil -destination x/chainlet/testutil/expected_keepers_mocks.go
