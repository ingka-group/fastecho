// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
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

import "fmt"

// printBanner prints a colored startup section with a title and alternating key/value pairs.
// Usage: printBanner("title", "key1", "value1", "key2", "value2")
func printBanner(title string, kvs ...string) {
	fmt.Printf("\n\033[1;36m⚡ %s\033[0m\n", title)
	for i := 0; i < len(kvs)-1; i += 2 {
		fmt.Printf("  \033[1;33m%-15s\033[0m : \033[1;32m%s\033[0m\n", kvs[i], kvs[i+1])
	}
}
