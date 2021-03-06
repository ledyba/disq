.PHONY: all run test bench get clean real-test

all: .bin/disq;

REPO=github.com/ledyba/disq
SRCS=$(shell find . -type f -name '*.go')

.bin/disq: .bin $(SRCS)
	gofmt -w .
	go generate "$(REPO)/cmd/disq"
	go build -o .bin/disq "$(REPO)/cmd/disq"

.bin/disq.linux: .bin $(SRCS)
	gofmt -w .
	go generate "$(REPO)/cmd/disq"
	GOOS=linux GOARCH=amd64 go build -o .bin/disq.linux "$(REPO)/cmd/disq"

.bin:
	mkdir -p .bin

run: all
	.bin/disq -v --config ./config-sample.json

test:
	go test "$(REPO)/..."

real-test: .bin/disq.linux
	scp .bin/disq.linux 202.229.192.118:~/disq
	scp config-test.json 202.229.192.118:~/config.json
	scp .bin/disq.linux 202.229.192.119:~/disq
	scp config-test.json 202.229.192.119:~/config.json

bench:
	go test -bench . "$(REPO)/..."

get:
	go get -u "github.com/Sirupsen/logrus"
	go get -u "github.com/fatih/color"
	go get -u "github.com/miekg/dns"
	go get -u "github.com/krolaw/dhcp4"

clean:
	go clean "$(REPO)/..."
	rm -rf .bin
