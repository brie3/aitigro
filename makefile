port ?= 8080
title ?= github_seeker

h help:
	@echo "h help - Instructions"
	@echo "build port=<number> title=<name> - builds package with specified port and name"
	@echo "run port=<number> runs package with specified port"
PHONY: h help

build:
	sed -i "s,__PORT__,$(port),g" server/server.go
	sed -i "s,__PORT__,$(port),g" web/index.html
	go build -o $(title)
PHONY: build

run:
	sed -i "s,__PORT__,$(port),g" server/server.go
	sed -i "s,__PORT__,$(port),g" web/index.html
	go run .
PHONY: run