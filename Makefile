DROPBOX=~/Dropbox/Public/geodns/dnsmonitor

all: templates

test: templates
	go test

dir: $(DROPBOX)

templates: data.go

data.go: templates/* static/* templates/*/* static/*/*
	sh bundle.sh

$(DROPBOX):
	mkdir -p $(DROPBOX)

sh: $(DROPBOX)/install.sh

setup: dir sh

linux: setup $(DROPBOX)/dnsmonitor-linux-x86_64 instructions

freebsd: setup $(DROPBOX)/dnsmonitor-freebsd-amd64 instructions

$(DROPBOX)/dnsmonitor-linux-x86_64:
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(DROPBOX)/dnsmonitor-linux-x86_64

$(DROPBOX)/dnsmonitor-freebsd-amd64:
	GOOS=freebsd \
	GOARCH=amd64 \
	go build -o $(DROPBOX)/dnsmonitor-freebsd-amd64

instructions:
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/dnsmonitor/install.sh | sh"

freebsd: setup
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(DROPBOX)/dnsmonitor-linux-x86_64
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/dnsmonitor/install.sh | sh"

$(DROPBOX)/install.sh: install.sh
	cp -p install.sh $(DROPBOX)/

push:
	( cd $(DROPBOX); sh ../push )
