// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema_test

import (
	"time"

	"github.com/juju/schema"
	gc "gopkg.in/check.v1"
)

type timeDurationSuite struct{}

var _ = gc.Suite(&timeDurationSuite{})

func (s *timeDurationSuite) TestTimeDuration(c *gc.C) {
	sch := schema.TimeDuration()

	var empty time.Duration

	out, err := sch.Coerce("", aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, empty)

	value, _ := time.ParseDuration("18h")

	out, err = sch.Coerce("18h", aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, value)

	out, err = sch.Coerce("failure", aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: conversion to duration: time: invalid duration \"failure\"")
	c.Check(out, gc.IsNil)

	out, err = sch.Coerce(42, aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Duration, got int(42)")
	c.Check(out, gc.IsNil)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Duration, got nothing")
	c.Check(out, gc.IsNil)
}

func (s *timeDurationSuite) TestTimeDurationString(c *gc.C) {
	sch := schema.TimeDurationString()

	out, err := sch.Coerce("", aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "0s")

	out, err = sch.Coerce("18h", aPath)
	c.Assert(err, gc.IsNil)
	// We get the long form because it's hours are greater than seconds.
	c.Check(out, gc.Equals, "18h0m0s")

	out, err = sch.Coerce("42m", aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "42m0s")

	out, err = sch.Coerce("42s", aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "42s")

	out, err = sch.Coerce("42ms", aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "42ms")

	out, err = sch.Coerce("failure", aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: conversion to duration: time: invalid duration \"failure\"")
	c.Check(out, gc.Equals, "")

	out, err = sch.Coerce(42, aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Duration, got int(42)")
	c.Check(out, gc.Equals, "")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Duration, got nothing")
	c.Check(out, gc.Equals, "")
}
