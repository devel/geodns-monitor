package main

import (
	"time"

	. "gopkg.in/check.v1"
)

type DurationSuite struct {
}

var _ = Suite(&DurationSuite{})

func (s *DurationSuite) TestDayDuration(c *C) {

	d := DayDuration{time.Second * 5}
	c.Check(d.DayString(), Equals, "5s")

	d = DayDuration{time.Second * 3670}
	c.Check(d.DayString(), Equals, "1h 1m")

}
