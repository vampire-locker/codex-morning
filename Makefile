APP := codex-morning
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin

.PHONY: build test install uninstall clean

build:
	go build -o bin/$(APP) ./cmd/$(APP)

test:
	go test ./...

install: build
	install -d "$(BINDIR)"
	install bin/$(APP) "$(BINDIR)/$(APP)"

uninstall:
	rm -f "$(BINDIR)/$(APP)"

clean:
	rm -rf bin
