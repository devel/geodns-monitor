"use strict";

var dns   = require('native-dns'),
      _   = require("underscore"),
timeago   = require('timeago'),
WebSocket = require('ws'),
boot_time = new Date().getTime();

var INTERVAL = 3000;
var SANITIZE_INTERVAL = 2000;

var config = { servers: {}, ws: {} };

var sanitize_timer;
var _sanitize_status = function() {
    if (!sanitize_timer) {
        sanitize_timer = setInterval(_sanitize_status, SANITIZE_INTERVAL);
    }
    _.each( _.keys(config.servers), function(ip) {
        var c = config.servers[ip];

        // don't trust the data if it doesn't update
        if (!c.timestamp
            || (c.timestamp &&
                (c.timestamp.getTime() + (INTERVAL * 3.5)) < new Date().getTime())
            ) {
            c.qps = 0;
            c.queries = 0;
            c.response_time = false;

            if (c.ws) {
                if (config.ws[ip]) {
                    config.ws[ip].close()
                    delete config.ws[ip];
                }
                c.ws = false;
            }
        }
    });
};

var _process_json = function(c) {

    var data = JSON.parse(c.json);
    if (data) {
        c.json = data;
    }
    var now = new Date();

    // pgeodns reset counts, so reset here too
    if (c.queries && data.qs < c.queries) {
        c.queries = 0;
    }

    if (c.queries && c.timestamp && data.qs) {
        var interval = (now - c.timestamp) / 1000;
        c.qps = parseInt(( data.qs - c.queries ) / interval, 10);
    }

    c.timestamp = now;
    if (data.qs) {
        c.queries = data.qs;
    }

    if (data.v) {
        c.version = data.v;
        if (c.version.indexOf(",") > 0) {
            c.version = c.version.slice(c.version.indexOf(",")+2);
        }
    }

    if (data.up) {
        c.uptime = data.up;
        c.uptime_p = timeago(new Date() - c.uptime * 1000);
    }

    c.status = "";

}

var _check_server = function(ip) {
    var c = config.servers[ip];
    if (!c) {
        console.error("Don't have configuration for ip", ip);
        return;
    }
    if (c.ws) {
        // have an active WS connection
        return;
    }
    var now = new Date();

    if (!c.ws_last_check || (now-c.ws_last_check > 90000)) {
        // console.log("time since last ws check", now-c.ws_last_check);
        c.ws_last_check = now;
        console.log("Trying WS for", ip);
        _check_server_ws(ip);
        return;
    }

    // no active WS connection, so use DNS
    _check_server_dns(ip);
};

var _check_server_ws = function(ip) {
    var c = config.servers[ip];
    if (c.ws) {
        // already connected
        return;
    }

    var ws = new WebSocket('ws://' + ip + ':8053/monitor',
        {origin: "http://dns-status.pgeodns"});

    c.ws = true;
    config.ws[ip] = ws;

    ws.on('error', function(err) {
        console.log("WS Error", c.ip, err);
        delete c.ws;
        c.status = err.code ? err.code : err;
        _check_server(ip);
    });

    ws.on('open', function(err) {
        if (err) {
            console.log("open err", err);
        }
        console.log("opened", c.ip);
        // ws.send('something');
    });

    ws.on('close', function() {
        console.log('disconnected', c.ip);
    });

    ws.on('message', function(data, flags) {
        //console.log(c.ip, data);
        c.json = data;
        c.status = "ws";        
        delete c.response_time;
        c.connection_type = "ws";
        _process_json(c);
    });

};

var _check_server_dns = function(ip) {
    var c = config.servers[ip];
    
    // console.log("supposed to check", c.ip, "last check", c.timestamp);

    c.check_start = new Date();
    c.connection_type = 'd';

    if (c.waiting) {
        // console.error("Already have a pending request for", c);
        c.status = "waiting";
        
        var time_limit = boot_time;
        if (c.timestamp) {
            time_limit = c.timestamp.getTime();
        }
        time_limit += INTERVAL * 6;
        // console.log("time limit", time_limit, "now:", new Date().getTime());
        
        if (time_limit < new Date().getTime()) {
            console.log("clearing waiting flag", ip);
            c.status  = "Retrying";
            c.waiting = false;
        }
        return;
    }
    else {
        c.waiting = true;
    }

    var question = dns.Question({
      name: '_status.pgeodns',
      type: dns.consts.NAME_TO_QTYPE.TXT
    });


    var req = dns.Request({
      question: question,
      server: { address: c.ip, port: 53, type: 'udp' },
      timeout: (INTERVAL * 1.8)
    });

    req.on('timeout', function () {
        c.json = "";
        c.status = "timeout";
        console.log('Timeout in making request', ip);
    });
    
    req.on('message', function (err, answer) {
        c.json = "";

        answer.answer.forEach(function (txt) {
            txt = txt.promote().data;
            //console.log("got txt record from ", c.ip, ":", txt);
            if (c.json) { c.json += "\n"; }
            c.json += txt;
        });
        return _process_json(c);
    });

    req.on('end', function () {
        c.waiting = false;
        if (c.json) {
            c.response_time = new Date() - c.check_start;
        }
        else {
            c.response_time = 0;
        }
        delete c.check_start;
    });

    req.send();
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
            console.log("New C IS", c);
            c.ip = a;
            //console.log("FQDN", fqdn);
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
                _check_server(c.ip);
                c.timer = setInterval(_check_server, INTERVAL, c.ip);
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
	if (!result) { console.error("Did not get results for", txt); }
        console.log("result", result);
        var names = result[0].split(" ");
        console.log("names", names);
        _.each(names, function(name) {
            var fqdn = name + "." + domain;
            _add_server(fqdn);
        });
    });
};

var add_server_by_name = function(name) {
    _add_server(name);
}

_sanitize_status();

module.exports = function() {
    return {
        add_servers_by_ns: add_servers_by_ns,
        add_servers_by_txt: add_servers_by_txt,
        add_server_by_name: add_server_by_name,
        status: function() {
            var r = _.clone(config);
            delete r.ws;
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


