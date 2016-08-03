// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

import (
	"fmt"
	"reflect"
)

// Omit is a marker for FieldMap and StructFieldMap defaults parameter.
// If a field is not present in the map and defaults to Omit, the missing
// field will be ommitted from the coerced map as well.
var Omit omit

type omit struct{}

type Fields map[string]Checker
type Defaults map[string]interface{}

// Schema represents a validator for a collection of related named attributes.
type Schema struct {
	// Checkers contains one validator for each attribute in the schema.
	Checkers map[string]Checker
	// Defaults contains the default value for an attribute if it is not set.
	Defaults map[string]interface{}
	// Dependencies defines an attribute whose existence depends on the exitence
	// and value of another attribute.
	Dependencies map[string]Dependency
	// Strict, if true, causes Coerce to return an error if an attribute is
	// passed in that is not in Checkers.
	Strict bool
}

// Dependency describes a dependenecy between two fields.  Currently, all
// dependencies are if and only if... the existence of the dependency with the
// given value means the named field must exist.  If the dependency does not
// exist or has a different value, the named field must not exist.
type Dependency struct {
	// DependsOn contains the name of the field that is depended on.
	DependsOn string `json:"name"`

	// Value holds the value that indicates this attribute is required.  Any
	// other value indicates the attribute should not be specified.
	Value interface{} `json:"value"`
}

// FieldMap returns a Checker that accepts a map value with defined
// string keys. Every key has an independent checker associated,
// and processing will only succeed if all the values succeed
// individually. If a field fails to be processed, processing stops
// and returns with the underlying error.
//
// Fields in defaults will be set to the provided value if not present
// in the coerced map. If the default value is schema.Omit, the
// missing field will be omitted from the coerced map.
//
// The coerced output value has type map[string]interface{}.
func FieldMap(checkers map[string]Checker, defaults map[string]interface{}) Checker {
	return Schema{
		Checkers: checkers,
		Defaults: defaults,
	}
}

// StrictFieldMap returns a Checker that acts as the one returned by FieldMap,
// but the Checker returns an error if it encounters an unknown key.
func StrictFieldMap(checkers map[string]Checker, defaults map[string]interface{}) Checker {
	return Schema{
		Checkers: checkers,
		Defaults: defaults,
		Strict:   true,
	}
}

var stringType = reflect.TypeOf("")

func hasStrictStringKeys(rv reflect.Value) bool {
	if rv.Type().Key() == stringType {
		return true
	}
	if rv.Type().Key().Kind() != reflect.Interface {
		return false
	}
	for _, k := range rv.MapKeys() {
		if k.Elem().Type() != stringType {
			return false
		}
	}
	return true
}

// Coerce validates the values in v by using the checkers in schema.Checkers to
// validate each value and convert it into a canonical form.  If unset and an
// entry in s.Defaults exists, that value will be set in the returned map.  v
// must be a map[string]interface{}, the result is a map[string]interface{} as
// well.
func (s Schema) Coerce(v interface{}, path []string) (interface{}, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, error_{"map", v, path}
	}
	if !hasStrictStringKeys(rv) {
		return nil, error_{"map[string]", v, path}
	}

	if s.Strict {
		for _, k := range rv.MapKeys() {
			name := k.String()
			if _, ok := s.Checkers[name]; !ok {
				return nil, fmt.Errorf("%sunknown key %q (value %#v)", pathAsPrefix(path), name, rv.MapIndex(k).Interface())
			}
		}
	}

	vpath := append(path, ".", "?")

	out := make(map[string]interface{}, rv.Len())
	for name, checker := range s.Checkers {
		valuev := rv.MapIndex(reflect.ValueOf(name))
		var value interface{}

		dflt, hasDefault := s.Defaults[name]
		_, hasDependency := s.Dependencies[name]
		switch {
		case valuev.IsValid():
			value = valuev.Interface()
		case hasDefault:
			if dflt == Omit {
				continue
			}
			value = dflt
		case hasDependency:
			// If an attribute has a dependency, it implicitly has a default of
			// Omit.  Whether or not it's ok for it not to exist in this
			// instance will be determined below.
			continue
		}
		vpath[len(vpath)-1] = name
		newv, err := checker.Coerce(value, vpath)
		if err != nil {
			return nil, err
		}
		out[name] = newv
	}
	for name, val := range s.Defaults {
		if val == Omit {
			continue
		}
		if _, ok := out[name]; !ok {
			checker, ok := s.Checkers[name]
			if !ok {
				return nil, fmt.Errorf("got default value for unknown field %q", name)
			}
			vpath[len(vpath)-1] = name
			newv, err := checker.Coerce(v, vpath)
			if err != nil {
				return nil, err
			}
			out[name] = newv
		}
	}

	// Dependencies are evaluated after coercion of values and after defaults
	// have been applied.

	for name, dep := range s.Dependencies {
		actual, depExists := out[dep.DependsOn]
		_, fieldExists := out[name]
		equal := reflect.DeepEqual(actual, dep.Value)
		switch {
		case !fieldExists && !depExists:
			// neither exist, that's fine.
			continue
		case fieldExists && depExists && equal:
			// both exist, and dependenecy has the correct value, that's fine.
			continue
		case fieldExists && depExists && !equal:
			return nil, fmt.Errorf("field %q should not be specified when %q is %v. %q requires %q to have value %v", name, dep.DependsOn, actual, name, dep.DependsOn, dep.Value)
		case fieldExists && !depExists:
			return nil, fmt.Errorf("field %q requires value %v be specified for %q, but it is not specified", name, dep.Value, dep.DependsOn)
		case !fieldExists && depExists && equal:
			return nil, fmt.Errorf("field %q exists with value %v, but required field %q is missing", dep.DependsOn, actual, name)
		}
	}
	return out, nil
}

// FieldMapSet returns a Checker that accepts a map value checked
// against one of several FieldMap checkers.  The actual checker
// used is the first one whose checker associated with the selector
// field processes the map correctly. If no checker processes
// the selector value correctly, an error is returned.
//
// The coerced output value has type map[string]interface{}.
func FieldMapSet(selector string, maps []Checker) Checker {
	fmaps := make([]Schema, len(maps))
	for i, m := range maps {
		if fmap, ok := m.(Schema); ok {
			if checker, _ := fmap.Checkers[selector]; checker == nil {
				panic("FieldMapSet has a FieldMap with a missing selector")
			}
			fmaps[i] = fmap
		} else {
			panic("FieldMapSet got a non-FieldMap checker")
		}
	}
	return mapSetC{selector, fmaps}
}

type mapSetC struct {
	selector string
	fmaps    []Schema
}

func (c mapSetC) Coerce(v interface{}, path []string) (interface{}, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, error_{"map", v, path}
	}

	var selector interface{}
	selectorv := rv.MapIndex(reflect.ValueOf(c.selector))
	if selectorv.IsValid() {
		selector = selectorv.Interface()
		for _, fmap := range c.fmaps {
			_, err := fmap.Checkers[c.selector].Coerce(selector, path)
			if err != nil {
				continue
			}
			return fmap.Coerce(v, path)
		}
	}
	return nil, error_{"supported selector", selector, append(path, ".", c.selector)}
}
