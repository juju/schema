// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

import (
	"strings"
)

// pathAsString returns a string consisting of the path elements. If path
// starts with a ".", the dot is omitted.
func pathAsString(path []string) string {
	if len(path) == 0 {
		return ""
	}
	if path[0] == "." {
		return strings.Join(path[1:], "")
	} else {
		return strings.Join(path, "")
	}
}
