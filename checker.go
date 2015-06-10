// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

// The Coerce method of the Checker interface is called recursively when
// v is being validated.  If err is nil, newv is used as the new value
// at the recursion point.  If err is non-nil, v is taken as invalid and
// may be either ignored or error out depending on where in the schema
// checking process the error happened. Checkers like OneOf may continue
// with an alternative, for instance.
type Checker interface {
	Coerce(v interface{}, path []string) (newv interface{}, err error)
}

// Any returns a Checker that succeeds with any input value and
// results in the value itself unprocessed.
func Any() Checker {
	return anyC{}
}

type anyC struct{}

func (c anyC) Coerce(v interface{}, path []string) (interface{}, error) {
	return v, nil
}

// OneOf returns a Checker that attempts to Coerce the value with each
// of the provided checkers. The value returned by the first checker
// that succeeds will be returned by the OneOf checker itself.  If no
// checker succeeds, OneOf will return an error on coercion.
func OneOf(options ...Checker) Checker {
	return oneOfC{options}
}

type oneOfC struct {
	options []Checker
}

func (c oneOfC) Coerce(v interface{}, path []string) (interface{}, error) {
	for _, o := range c.options {
		newv, err := o.Coerce(v, path)
		if err == nil {
			return newv, nil
		}
	}
	return nil, error_{"", v, path}
}
