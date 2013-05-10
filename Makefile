all: templates dnsmonitor

debug:
	go build -ldflags "-s"  -gcflags "-N -l"

gdb: debug
	gdb ./dnsmonitor -d $GOROOT

dnsmonitor: *.go
	go build

templates: static/js/templates.js

static/js/templates.js: templates/client/*
	(cd templates/client && hulk *.html > ../../static/js/templates.js)
