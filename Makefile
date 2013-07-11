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

linux: dir sh
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(DROPBOX)/dnsmonitor-linux-x86_64
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/dnsmonitor/install.sh | sh"

$(DROPBOX)/install.sh: dir install.sh
	cp install.sh $(DROPBOX)/

push:
	( cd $(DROPBOX); sh ../push )
