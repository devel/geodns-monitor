package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "gopkg.in/check.v1"
)

type HTTPSuite struct {
	hub *StatusHub
}

var _ = Suite(&HTTPSuite{})

func (s *HTTPSuite) SetUpSuite(c *C) {
	fmt.Println("Starting http server")
	hub := NewHub()
	go startHTTP(6824, hub)
	time.Sleep(20 * time.Millisecond)
}

func (s *HTTPSuite) TestSetup(c *C) {
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
