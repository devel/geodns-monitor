package main

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

func TestConfig(t *testing.T) {
	hub := NewHub()
	configure(hub)

}
