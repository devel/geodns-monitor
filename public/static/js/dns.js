"use strict";

(function ($) {
    
    var current_popover;
    
    var update = function() {
        $.getJSON('/api/status', function(status) {
            //console.log("c", status);
            var servers = status.servers; 
            // _.filter(status.servers, function(s) { return s.status.match("1.40") })
            
            $('#servers span[rel=tooltip]').tooltip('hide');
            $('#servers tbody').html("");
            _.each( _.sortBy(servers, function(s) { return s.name }), function(s) {
                s.names = _.map(s.names, function(n) { return { name: n } });
                s.qps_class = s.qps && s.qps > 150 ? "high-query-rate" : "";
                s.response_time_class = (s.response_time && s.response_time > 400) ? "slow-response" : "";
                var template = templates.server.render({ server: s });
                $('#servers').append(template);
            });
            $('#summary').html( templates.summary.render( status ) );
            $('#servers span[rel=tooltip]').tooltip({trigger: "hover", placement: "right"});

            var str = JSON.stringify(status, undefined, 2);
            $('#status_dump').html(str);
        });
    };

    $('#debug_toggle').on('click', function(e) {
        $('#status_dump').toggle();
    });

    $('#servers').on('click', "a.ip", function(e) {
        e.preventDefault();
        var ip = $(this).text();
        console.log("current_popover", current_popover);
        console.log("clicked on ip", ip, this) ;
        if (current_popover) { current_popover.popover('hide') }
        $(this).popover({ trigger: "manual", "title": "foo" });
        $(this).popover("show");
        current_popover = $(this);
        
    });

    /*
    $('#servers span.ip').popover({
        content: function() {
            var ip = $(this).text();
            return '<pre>'
                + JSON.stringify(status.servers[ip], undefined, 2)
                + '</pre>';
        }
    });
    */

    // popover demo
    $("a[rel=popover]")
      .popover()
      .click(function(e) {
        e.preventDefault()
      })


    update();
    window.setInterval(update, 3000);
})(jQuery);
