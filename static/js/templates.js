if (!!!templates) var templates = {};
templates["server"] = new Hogan.Template({code: function (c,p,i) { var t=this;t.b(i=i||"");t.b("<tr>");t.b("\n" + i);if(t.s(t.f("server",c,p,1),c,p,0,16,538,"{{ }}")){t.rs(c,p,function(c,p,t){t.b("<td><span style=\"background-color:rgb(");t.b(t.v(t.f("color",c,p,0)));t.b(")\">&nbsp;</span> <span rel=\"tooltip\" title=\"");if(t.s(t.f("names",c,p,1),c,p,0,118,127,"{{ }}")){t.rs(c,p,function(c,p,t){t.b(t.v(t.f("name",c,p,0)));t.b(" ");});c.pop();}t.b("\">");t.b(t.v(t.f("name",c,p,0)));t.b("</span></td>");t.b("\n");t.b("\n" + i);t.b("<td>");if(t.s(t.f("Data",c,p,1),c,p,0,174,191,"{{ }}")){t.rs(c,p,function(c,p,t){t.b(t.v(t.f("connection_id",c,p,0)));});c.pop();}t.b("</td>");t.b("\n");t.b("\n" + i);t.b("<td><span class=\"ip\">");t.b("\n" + i);t.b("<a href=\"http://");t.b(t.v(t.f("ip",c,p,0)));t.b(":8053/status\">");t.b(t.v(t.f("ip",c,p,0)));t.b("</a>");t.b("\n" + i);t.b("</td>");t.b("\n");t.b("\n" + i);t.b("<td class=\"");t.b(t.v(t.f("qps_class",c,p,0)));t.b("\">");t.b("\n" + i);t.b("    ");if(t.s(t.f("qps",c,p,1),c,p,0,322,333,"{{ }}")){t.rs(c,p,function(c,p,t){t.b(t.v(t.f("qps",c,p,0)));t.b("/qps");});c.pop();}t.b("\n" + i);t.b("</td>");t.b("\n");t.b("\n" + i);t.b("<td>");t.b("\n" + i);t.b("	");if(t.s(t.f("qps1m",c,p,1),c,p,0,365,378,"{{ }}")){t.rs(c,p,function(c,p,t){t.b(t.v(t.f("qps1m",c,p,0)));t.b("/qps");});c.pop();}t.b("\n" + i);t.b("</td>");t.b("\n");t.b("\n" + i);t.b("<td>");t.b(t.v(t.f("version",c,p,0)));t.b("</td>");t.b("\n" + i);t.b("<td><small>");if(t.s(t.f("groups",c,p,1),c,p,0,439,445,"{{ }}")){t.rs(c,p,function(c,p,t){t.b(t.v(t.d(".",c,p,0)));t.b(" ");});c.pop();}t.b("</small></td>");t.b("\n" + i);t.b("<td>");t.b(t.v(t.f("uptime_p",c,p,0)));t.b("</td>");t.b("\n" + i);t.b("<td>");t.b(t.v(t.f("last_update",c,p,0)));t.b("</td>");t.b("\n" + i);t.b("<td>");t.b(t.v(t.f("status",c,p,0)));t.b("</td>");t.b("\n");t.b("\n" + i);});c.pop();}t.b("</tr>");return t.fl(); },partials: {}, subs: {  }});
templates["summary"] = new Hogan.Template({code: function (c,p,i) { var t=this;t.b(i=i||"");if(t.s(t.f("summary",c,p,1),c,p,0,12,88,"{{ }}")){t.rs(c,p,function(c,p,t){t.b("    <span class=\"btn btn-info btn-large\">");t.b(t.v(t.f("qps",c,p,0)));t.b(" queries per second</span>");t.b("\n" + i);});c.pop();}return t.fl(); },partials: {}, subs: {  }});
