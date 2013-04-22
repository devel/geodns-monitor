package main

import (
	. "launchpad.net/gocheck"
)

type StatusHubSuite struct {
	hub *StatusHub
}

var _ = Suite(&StatusHubSuite{})

func (s *StatusHubSuite) SetUpSuite(c *C) {
	s.hub = NewHub()
}

func (s *StatusHubSuite) TestHub(c *C) {

	c.Log("Starting Hub test")

	err := s.hub.AddName("abc")
	c.Check(err, ErrorMatches, "Could not lookup name:.*")

	err = s.hub.AddName("127.0.0.1")
	c.Check(err, IsNil)

	s.hub.Stop()
}
