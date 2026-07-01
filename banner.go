// Copyright © 2026 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fastecho

import (
	"fmt"
	"runtime/debug"
)

// modulePath is fastecho's module path, used to find its version in the
// consuming binary's build info.
const modulePath = "github.com/ingka-group/fastecho"

// version reports fastecho's module version from the consuming binary's build
// info, or "dev" when unavailable (e.g. running from this repo's own tree).
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath && dep.Version != "" {
			return dep.Version
		}
	}
	return "dev"
}

// printBanner prints a startup section with a title and alternating key/value pairs.
// Keys align to a 16-col gutter (wider when a key exceeds it) so sections line up.
// Usage: printBanner("title", "key1", "value1", "key2", "value2")
func printBanner(title string, kvs ...string) {
	width := 16
	for i := 0; i+1 < len(kvs); i += 2 {
		if len(kvs[i]) > width {
			width = len(kvs[i])
		}
	}
	fmt.Printf("\n%s\n", title)
	for i := 0; i+1 < len(kvs); i += 2 {
		fmt.Printf("  %-*s : %s\n", width, kvs[i], kvs[i+1])
	}
}
