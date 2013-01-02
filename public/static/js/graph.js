/**
 * Manages the SmoothieCharts library with the two graphs we are generating on the index page
 * http://smoothiecharts.org
 */
var graph = (function () {
    "use strict";

    var smoothieServers = new SmoothieChart(), // One line per server chart
        timeSeries      = {},
        smoothieTotal   = new SmoothieChart(), // Summary chart
        timeSeriesTotal = new TimeSeries(),
        serverColors    = {},
        colors          = [];

    smoothieServers.streamTo(document.getElementById("graphServers"), 1000 /*delay*/);
    smoothieTotal.streamTo(document.getElementById("graphTotal"), 1000 /*delay*/);
    smoothieTotal.addTimeSeries(timeSeriesTotal);

    return {
        "record": function (serverName, qps) {
            if (timeSeries[serverName] === undefined) {
                timeSeries[serverName]   = new TimeSeries();
                var lineColor            = colors.shift() || [255, 255, 255];
                serverColors[serverName] = lineColor;
                smoothieServers.addTimeSeries(timeSeries[serverName], {
                    strokeStyle: 'rgb(' + lineColor.join(',') + ')'
                });
            }
            if (serverName === "summary") {
                timeSeriesTotal.append(new Date().getTime(), parseInt(qps, 10));
            } else {
                timeSeries[serverName].append(new Date().getTime(), parseInt(qps, 10));
            }
        },
        "generateColors": function (n) {
            if (colors.length === 0) {
                colors = generateColors(n);
            }
        },
        "getColor": function (serverName) {
            return serverColors[serverName];
        }
    };
}());
