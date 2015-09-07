package main

import (
	. "gopkg.in/check.v1"

	"time"
)

type StatusHubSuite struct {
	hub *StatusHub
}

var _ = Suite(&StatusHubSuite{})

func (s *StatusHubSuite) SetUpSuite(c *C) {
	s.hub = NewHub()
}

func (s *StatusHubSuite) TestHub(c *C) {

	err := s.hub.AddName("abc")
	c.Check(err, ErrorMatches, "Could not lookup name:.*")

	err = s.hub.AddName("127.0.0.1")
	c.Check(err, IsNil)

	c.Check(s.hub.Status(), HasLen, 1)

	s.hub.MarkConfigurationStart()

	err = s.hub.AddName("127.0.0.2")
	c.Check(err, IsNil)

	s.hub.MarkConfigurationEnd()

	time.Sleep(3 * time.Second)

	statuses := s.hub.Status()
	c.Check(statuses[0].Status, Equals, "stopped")
	c.Check(statuses[1].Status, Equals, "Starting")

	s.hub.Stop()
}
