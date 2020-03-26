BINARY_NAME := ov
SRCS := $(shell git ls-files '*.go')

all: build

test: $(SRCS)
	go test ./...

build: $(BINARY_NAME)

$(BINARY_NAME): $(SRCS)
	go build -o $(BINARY_NAME) ./cmd/ov

install:
	go install ./cmd/ov

clean:
	rm -f $(BINARY_NAME)

.PHONY: all test build install clean
