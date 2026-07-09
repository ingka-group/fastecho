// Copyright © 2026 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package banner prints the startup banner sections, so every section of the
// boot output shares one format and column width.
package banner

import "fmt"

// Section prints a titled key/value block on stdout. kvs alternate key, value.
func Section(title string, kvs ...string) {
	fmt.Printf("\n%s\n", title)
	for i := 0; i+1 < len(kvs); i += 2 {
		fmt.Printf("  %-22s : %s\n", kvs[i], kvs[i+1])
	}
}
