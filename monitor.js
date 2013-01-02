#!/usr/bin/env node
"use strict";

/*
TODO:

* If the uptime drops, clear the previous query count
* Periodically poll for a new server lists. How do we purge old servers? Have
  to keep track of where the servers came from, or syncronize the config update.
* Check SOA timestamps
* Graph query rates
* Stream updates with socket.io
* "Score chart" of top query rates
* cube emitter
* nagios check URLs
   /nagios/monitor - check monitor itself
   /nagios/server  - all servers, warn if any are broken, red if more than X%
   /nagios/server/IP - status of that IP
* automatically generate nagios config?

*/

var express = require('express'),
    cons    = require('consolidate'),
    app     = express(),
    dnsmonitor = require('./lib/dns-monitor')(),
    fs      = require('fs'),
    hogan   = require('hulk-hogan'),
    panic   = require('panic');
var Monitor = {};
Monitor.PACKAGE = (function() {
    var json = fs.readFileSync(__dirname + '/package.json', 'utf8');
    return JSON.parse(json);
}());

app.configure(function() {
    app.set('views', __dirname + '/views');
    app.engine('html', cons.hogan);
    app.use(express.bodyParser());
    app.use(app.router);
});

app.use(express['static'](__dirname + "/public"));
//app.use(require('connect-assets'));

app.get('/', function(req, res){
    res.locals.version = Monitor.PACKAGE.version;
    res.render('index.html');
});

app.get('/api/status', function(req,res) {
    var config = dnsmonitor.status();
    res.header('Cache-Control', 'max-age=1');
    // console.log("config is", config);
    return res.json(config);
});

app.get('/api/status_events', function(req,res) {
    res.header('Content-Type', 'text/event-stream');
    res.header('Cache-Control', 'no-cache');
    res.header('Connection', 'keep-alive');

    dnsmonitor.events.on('status', function(status) {
        res.write('event: status\n');
        res.write('data: ' + JSON.stringify(status) + '\n\n');
    });

    res.on('close', function() {
        // fs.unwatchFile(logfile);
        // fs.close(fd);
    });
});

dnsmonitor.add_servers_by_ns("pool.ntp.org");
dnsmonitor.add_servers_by_txt("all-dns.ntppool.net");
dnsmonitor.add_servers_by_ns("android.ntppool.org");
dnsmonitor.add_servers_by_ns("cpansearch.perl.org");
dnsmonitor.add_server_by_name('zrh4.ntppool.net');
dnsmonitor.add_server_by_name('lax4a.ntppool.net');

var port = 1090;
console.log("listening to port", port);
app.listen(port);