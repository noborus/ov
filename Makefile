BINARY_NAME := zpager
SRCS := $(shell git ls-files '*.go')

all: build

test: $(SRCS)
	go test ./...

build: $(BINARY_NAME)

$(BINARY_NAME): $(SRCS)
	go build -o $(BINARY_NAME) ./cmd/zpager

install:
	go install ./cmd/zpager

clean:
	rm -f $(BINARY_NAME)

.PHONY: all test build install clean
