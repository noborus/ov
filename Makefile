BINARY_NAME := ov
SRCS := $(shell git ls-files '*.go')
LDFLAGS := "-X main.Version=$(shell git describe --tags --abbrev=0 --always) -X main.Revision=$(shell git rev-parse --verify --short HEAD)"

all: build

test: $(SRCS)
	go test ./...

build: $(BINARY_NAME)

$(BINARY_NAME): $(SRCS)
	go build -ldflags $(LDFLAGS) -o $(BINARY_NAME) ./cmd/ov

install:
	go install -ldflags $(LDFLAGS) ./cmd/ov

clean:
	rm -f $(BINARY_NAME)

.PHONY: all test build install clean
