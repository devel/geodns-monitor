#!/usr/bin/env node
"use strict";

/*
TODO:

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
    app = module.exports = express.createServer(),
    dnsmonitor = require('./lib/dns-monitor')(),
    panic = require('panic');

app.configure(function() {
	app.set('views', __dirname + '/views');
	app.register('.html', require('express-hogan.js'));
	app.use(express.bodyParser());
	app.use(app.router);
});

app.use(express['static'](__dirname + "/public"));
//app.use(require('connect-assets'));

app.get('/', function(req, res){
    res.render('index.html');
});

app.get('/api/status', function(req,res) {
    var config = dnsmonitor.status();
    res.header('Cache-Control', 'max-age=1');
    //console.log("config is", config);
    return res.json(config);
});

dnsmonitor.add_servers_by_ns("pool.ntp.org");
dnsmonitor.add_servers_by_txt("test.ntpns.org");
dnsmonitor.add_servers_by_txt("all-dns.ntppool.net");


var port = 1090;
console.log("listening to port", port);
app.listen(port);