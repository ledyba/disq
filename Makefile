.PHONY: all run test get clean

all: .bin/disq;

REPO=github.com/ledyba/disq
SRCS=$(shell find . -type f -name '*.go')

.bin/disq: .bin $(SRCS)
	gofmt -w .
	go generate "$(REPO)/cmd/disq"
	go build -o .bin/disq "$(REPO)/cmd/disq"

.bin:
	mkdir -p .bin

run: all
	.bin/disq -v --config ./config-sample.json

test:
	go test "$(REPO)/..."

get:
	go get -u "github.com/Sirupsen/logrus"
	go get -u "github.com/fatih/color"
	go get -u "github.com/miekg/dns"
	go get -u "github.com/krolaw/dhcp4"

clean:
	go clean "$(REPO)/..."
	rm -rf .bin
