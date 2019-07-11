.PHONY: clean

GOPATH=$(PWD)
GOCMD=GOPATH=$(GOPATH) go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BINARY_NAME=zsm
LINUX_386_NAME=$(BINARY_NAME)-linux-386
LINUX_AMD64_NAME=$(BINARY_NAME)-linux-amd64
LINUX_ARM5_NAME=$(BINARY_NAME)-linux-arm5
LINUX_ARM6_NAME=$(BINARY_NAME)-linux-arm6
LINUX_ARM7_NAME=$(BINARY_NAME)-linux-arm7
LINUX_ARM64_NAME=$(BINARY_NAME)-linux-arm64
DARWIN_386_NAME=$(BINARY_NAME)-darwin-386
DARWIN_AMD64_NAME=$(BINARY_NAME)-darwin-amd64
WINDOWS_386_NAME=$(BINARY_NAME)-windows-386
WINDOWS_AMD64_NAME=$(BINARY_NAME)-windows-amd64

all: build

build: main

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf pkg
	rm -rf bin
	
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

main:
	$(GOGET) -u github.com/atlaslee/zlog
	$(GOBUILD) zsm
	$(GOINSTALL) zsm