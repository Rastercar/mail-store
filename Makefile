.PHONY: all get build test clean run run_dev 
.DEFAULT_GOAL: $(BIN_FILE)

PROJECT_NAME = mail_store_ms

CMD_DIR = ./cmd

BIN_FILE = ./bin/$(PROJECT_NAME)
CONFIG_FILE = ./config/config.yml
DEV_CONFIG_FILE = ./config/config.dev.yml

# Get version constant
VERSION := $(shell git describe --abbrev=0 --tags --always)
BUILD := $(shell git rev-parse HEAD)

# Use linker flags to provide version/build settings to the binary
LDFLAGS=-ldflags "-s -w -X=main.version=$(VERSION) -X=main.build=$(BUILD)"

# Fetch dependencies
get:
	@echo "[*] Downloading dependencies..."
	cd $(CMD_DIR) && go get
	@echo "[*] Finish..."

# Build the go binary
build:
	@echo "[*] Building $(PROJECT_NAME)..."
	go build $(LDFLAGS) -o $(BIN_FILE) $(CMD_DIR)/...
	@echo "[*] Finish..."

# Run all tests
test:
	go test -race -cover -coverprofile=coverage.out ./... 
	go tool cover -html=coverage.out

# Clears all compiled content
clean:
	rm -rf bin/
	rm -rf vendor/

# Builds and runs the aplication on production mode
run: build
	$(BIN_FILE) -config-file=$(CONFIG_FILE)

# Builds and runs the aplication on development mode
run_dev: build
	$(BIN_FILE) -config-file=$(DEV_CONFIG_FILE)

# run the service as a container
docker_run:
	docker run mail_store_ms:latest --config-file="/etc/config.dev.yml"

# run the service as a container with the dev config file
docker_run_dev:
	docker run mail_store_ms:latest --config-file="/etc/config.dev.yml"

# build the service
docker_build:
	docker build -f ./docker/Dockerfile . -t mail_store_ms

# ------- DOCKERFILE SPECIFIC COMMANDS

# Copy the files necessary to execute the built binary
# (made to be used within a alpine docker image) 
install:
	mkdir -p /etc/$(PROJECT_NAME)/
	cp $(BIN_FILE) /usr/local/bin/
	cp $(CONFIG_FILE) /etc/
	cp $(DEV_CONFIG_FILE) /etc/

# Clears all the files necessary to execute the built binary
uninstall:
	rm -rf /usr/local/bin/$(BIN_FILE)
	rm -rf /etc/$(PROJECT_NAME)/