package main

import (
	. "launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

func TestConfig(t *testing.T) {
	hub := NewHub()
	configure(hub)

}
