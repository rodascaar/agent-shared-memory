# Pizarrón ligero — empaquetado mínimo (KISS: make basta, sin GoReleaser)
# CGO_ENABLED=0: binarios 100% estáticos (sin libc/Xcode) para distribuir.

VERSION ?= v0.1.0
BIN     := pizarron
PKG     := ./cmd/pizarron
LDFLAGS := -s -w -X main.Version=$(VERSION)
PORT    ?= 18081
PLATFORMS := darwin-arm64 darwin-amd64 linux-arm64 linux-amd64 windows-amd64
export CGO_ENABLED := 0

.PHONY: build build-all test test-e2e clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

build-all:
	mkdir -p dist
	for target in $(PLATFORMS); do \
		os=$${target%-*}; arch=$${target#*-}; \
		out=dist/$(BIN)-$${os}-$${arch}; \
		if [ "$$os" = "windows" ]; then out=$$out.exe; fi; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o $$out $(PKG) || exit 1; \
	done
	cd dist && (sha256sum $(BIN)-* 2>/dev/null || shasum -a 256 $(BIN)-*) > SHA256SUMS && cat SHA256SUMS

test:
	go vet ./...
	go test ./...

# E2E: levanta el binario local en un puerto efímero y corre el contrato universal.
test-e2e: build
	(PIZARRON_ADDR=127.0.0.1:$(PORT) PIZARRON_LOG=/tmp/pizarron-e2e.log ./$(BIN) & echo $$! > /tmp/pizarron-e2e.pid); \
	sleep 1; \
	PIZARRON_URL=http://127.0.0.1:$(PORT) sh api/contract.sh; code=$$?; \
	kill $$(cat /tmp/pizarron-e2e.pid) 2>/dev/null; rm -f /tmp/pizarron-e2e.log; \
	exit $$code

clean:
	rm -f $(BIN) pizarron.log
	rm -rf dist
