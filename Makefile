all: templates dnsmonitor

dnsmonitor: *.go
	go build

templates: static/js/templates.js

static/js/templates.js: templates/client/*
	(cd templates/client && hulk *.html > ../../static/js/templates.js)
