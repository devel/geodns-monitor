package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "gopkg.in/check.v1"
)

type HTTPSuite struct {
	hub *StatusHub
	srv *httptest.Server
}

var _ = Suite(&HTTPSuite{})

func (s *HTTPSuite) SetUpSuite(c *C) {
	fmt.Println("Starting http server")
	hub := NewHub()
	s.srv = httptest.NewServer(setupMux(hub))
}

func (s *HTTPSuite) TestSetup(c *C) {
	res, err := http.Get(s.srv.URL + "/")
	c.Assert(err, IsNil)
	page, _ := ioutil.ReadAll(res.Body)
	c.Check(string(page), Matches, "(?s).*<title>geodns monitor.*")

	// API working?
	res, err = http.Get(s.srv.URL + "/api/status")
	c.Assert(err, IsNil)
	c.Check(res.StatusCode, Equals, 200)

	// Fetch static files
	res, err = http.Get(s.srv.URL + "/static/js/dns.js")
	c.Assert(err, IsNil)
	c.Check(res.StatusCode, Equals, 200)
	page, _ = ioutil.ReadAll(res.Body)

	if s.hub != nil {
		s.hub.Stop()
	}
}
