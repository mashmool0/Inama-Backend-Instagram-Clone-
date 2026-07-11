BUF_VERSION := v1.47.0
BUF := $(shell command -v buf 2>/dev/null)

.PHONY: gen lint format buf-install

# One command to turn the .proto contracts into Go + Python code.
#   make gen
# Installs buf if missing, generates into proto/gen/, then tidies the
# generated Go module so its dependencies are resolved.
gen: buf-install
	buf generate
	cd proto/gen && go mod tidy
	@echo "✓ code generated in proto/gen/ — commit it"

# Fails if any .proto breaks the style rules. Run before pushing.
lint: buf-install
	buf lint

# Auto-formats all .proto files in place.
format: buf-install
	buf format -w

buf-install:
ifndef BUF
	@echo "buf not found — installing $(BUF_VERSION)..."
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	@echo "buf installed to $$(go env GOPATH)/bin — make sure that's on your PATH"
endif
