NAME=ovftool
BINARY=packer-plugin-${NAME}

build:
	@go build -o ${BINARY}
.PHONY: build

dev: build
	@mkdir -p ~/.packer.d/plugins/
	@mv ${BINARY} ~/.packer.d/plugins/${BINARY}
.PHONY: dev

generate:
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest
	@go generate -v ./...
.PHONY: generate
