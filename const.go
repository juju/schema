// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

import (
	"fmt"
	"reflect"
)

// Const returns a Checker that only succeeds if the input matches
// value exactly.  The value is compared with reflect.DeepEqual.
func Const(value interface{}) Checker {
	return constC{value}
}

type constC struct {
	value interface{}
}

func (c constC) Coerce(v interface{}, path []string) (interface{}, error) {
	if reflect.DeepEqual(v, c.value) {
		return v, nil
	}
	return nil, error_{fmt.Sprintf("%#v", c.value), v, path}
}

// Empty returns a Checker that only succeeds if the input is an empty value
// (nil). To tweak the error message, valueLabel can contain a label of the
// value being checked to be empty, e.g. "my special name". If valueLabel is "",
// "value" will be used as a label instead.
func Empty(valueLabel string) Checker {
	if valueLabel == "" {
		valueLabel = "value"
	}
	return emptyC{valueLabel}
}

type emptyC struct {
	valueLabel string
}

func (c emptyC) Coerce(v interface{}, path []string) (interface{}, error) {
	if reflect.DeepEqual(v, nil) {
		return v, nil
	}
	label := fmt.Sprintf("empty %s", c.valueLabel)
	return nil, error_{label, v, path}
}
