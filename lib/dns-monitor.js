"use strict";

var dns = require('native-dns'),
      _ = require("underscore"),
timeago = require('./timeago');

var config = { servers: {} };

var sanitize_timer;
var _sanitize_status = function() {
    if (!sanitize_timer) {
        sanitize_timer = setInterval(_sanitize_status, 10000);
    }
    _.each( _.keys(config.servers), function(ip) {
        var c = config.servers[ip];
        if (c.timestamp + 10000 < new Date()) {
            // don't trust the data if it doesn't update
            c.qps = 0;
            c.queries = 0;
        }
    });
};

var _check_server = function(ip) {
    var c = config.servers[ip];
    if (!c) {
        console.error("Don't have configuration for ip", ip);
    }
    console.log("supposed to check", c.ip, "last check", c.timestamp);

    c.check_start = new Date();

    if (c.waiting) {
        console.error("Already have a pending request");
        c.status = "Waiting for response";
        return;
    }
    else {
        c.waiting = true;
    }

    var request = dns.resolve('_status.pgeodns', 'TXT', c.ip,
        function (err, results) {
            var i;
            c.json = "";

            c.waiting = false;

            c.response_time = new Date() - c.check_start;
            delete c.check_start;

            if (c.timer) {
                console.log("already has a timer", c.timer);
                clearTimeout(c.timer);
            }
            c.timer = setTimeout(_check_server, 2000, c.ip);

            if (err) {
                console.log(err);
                c.status = "Error getting status: " + err;
                // we sometimes get timeouts and then the response
                // so only re-schedule the check on successful results
                // and on errors just add a later "maybe restart" event
                /* setTimeout(function() {
                    if (!c.timer) {
                        _check_server(c.ip);
                    }
                }, 10000)
                */
                return;
            }

            for (i = 0; i < results.length; i++) {
                console.log('results:', results[i]);
                if (c.json) { c.json += "\n"; }
                c.json += results[i];
            }
            var data = JSON.parse(c.json);
            if (data) {
                c.json = data;
            }
            var now = new Date();

            // pgeodns reset counts, so reset here too
            if (c.queries && data.qs < c.queries) {
                c.queries = 0;
            }

            if (c.queries && c.timestamp) {
                var interval = (now - c.timestamp) / 1000;
                c.qps = parseInt(( data.qs - c.queries ) / interval, 10);
            }

            c.timestamp = now;
            c.queries = data.qs;

            c.version = data.v;
            if (c.version.indexOf(",")) {
                c.version = c.version.slice(c.version.indexOf(",")+2);
            }

            c.uptime = data.up;
            c.uptime_p = timeago(new Date() - c.uptime * 1000);

            c.status = ""; // "Updated " + now.toUTCString();
            //if (c.qps) {
            //    c.status += " " + c.qps + "/qps";
            //}

        });
};

var _add_server = function(fqdn) {
    // TOOD: also check AAAA?
    dns.resolve(fqdn, "A", function(err, records) {
        console.log("looked up", fqdn, "got A:", records);
        _.each(records, function(a) {
            if (!config.servers[a]) {
                config.servers[a] = {
                    names: [],
                    version: "",
                    queries: 0,
                    status: ""
                };
            }
            var c = config.servers[a];
            console.log("C IS", c);
            c.ip = a;
            console.log("FQDN", fqdn);
            c.names.push(fqdn);
            c.names = _.uniq(c.names);
            c.name = _.reduce(c.names, function(memo, name) {
                var short = name.slice(0, name.indexOf("."));
                if (short.length > memo.length) {
                    return short;
                }
                else {
                    return memo;
                }
            }, "");
            if (!c.timer) {
                c.timer = true;
                _check_server(c.ip);
            }
        });
    });
};

var add_servers_by_ns = function(domain) {
    console.log("adding servers serving", domain);
    dns.resolveNs(domain, function(err, result) {
        _.each(result, function(ns) {
            _add_server(ns);
        });
    });
};

var add_servers_by_txt = function(txt, domain) {
    if (!domain) {
        domain = txt.slice(txt.indexOf(".")+1);
    }
    console.log("adding servers for", txt, "base domain", domain);
    dns.resolveTxt(txt, function(err, result) {
        if (err) { console.error("Could not resolve", txt, "error:", err); }
        console.log("result", result);
        var names = result[0].split(" ");
        console.log("names", names);
        _.each(names, function(name) {
            var fqdn = name + "." + domain;
            _add_server(fqdn);
        });
    });
};



module.exports = function() {
    return {
        add_servers_by_ns: add_servers_by_ns,
        add_servers_by_txt: add_servers_by_txt,
        status: function() {
            var r = _.clone(config);
            var summary = { qps: 0 };
            _.each( _.keys(r.servers), function(ip) {
                var c = r.servers[ip];
                c.last_update = timeago( c.timestamp );
                // TODO: make the anycast IP configurable via DNS
                if (c.qps && c.ip !== '207.171.17.42') {
                    summary.qps += c.qps;
                }
                delete c.timer;
            });
            r.summary = summary;
            return r;
        }
    };
};


