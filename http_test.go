package main

import (
	"fmt"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"net/http"
	"time"
)

type HttpSuite struct {
	hub *StatusHub
}

var _ = Suite(&HttpSuite{})

func (s *HttpSuite) SetUpSuite(c *C) {
	fmt.Println("Starting http server")
	hub := NewHub()
	go startHttp(6824, hub)
	time.Sleep(20 * time.Millisecond)
}

func (s *HttpSuite) TestSetup(c *C) {
	res, err := http.Get("http://localhost:6824/")
	c.Assert(err, IsNil)
	page, _ := ioutil.ReadAll(res.Body)
	c.Check(string(page), Matches, "(?s).*<title>geodns monitor.*")

	// Fetch static files
	res, err = http.Get("http://localhost:6824/static/js/dns.js")
	c.Assert(err, IsNil)
	c.Check(res.StatusCode, Equals, 200)
	page, _ = ioutil.ReadAll(res.Body)

	if s.hub != nil {
		s.hub.Stop()
	}
}
