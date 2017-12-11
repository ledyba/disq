.PHONY: all run get clean

all: .bin/disq;

.bin/disq:
	gofmt -w .
	go build -o .bin/disq "github.com/ledyba/disq/cmd/disq"

.bin:
	mkdir -p .bin

run: all
	.bin/disq

get:
	go get -u "github.com/Sirupsen/logrus"
	go get -u "github.com/fatih/color"
	go get -u "github.com/miekg/dns"

clean:
	go clean github.com/ledyba/embed-markdown/...
