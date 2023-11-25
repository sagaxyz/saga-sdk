all: proto

.PHONY: test
test:
	go test -race -cover -coverprofile cp.out -count=1 -timeout=30s ./...

.PHONY: proto proto-gen
proto: proto-gen

proto_ver=0.11.6
proto_image_name=ghcr.io/cosmos/proto-builder:$(protoVer)
proto_image=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)
proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh
