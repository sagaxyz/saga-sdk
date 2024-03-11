all: proto

.PHONY: test
test:
	go test -race -cover -coverprofile cp.out -count=1 -timeout=30s ./...

.PHONY: proto proto-gen
proto: proto-gen

proto_ver=latest
proto_image_name=ghcr.io/cosmos/proto-builder:$(proto_ver)
proto_image=docker run --rm -v $(CURDIR):/workspace --workdir /workspace $(proto_image_name)
proto-gen:
	@echo "Generating Protobuf files"
	@$(proto_image) sh ./scripts/protocgen.sh
