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

package stringutils

import (
	"strconv"
	"strings"
)

// IsEmpty checks whether the given string is empty
func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// ToInt converts a string to an int
func ToInt(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return -1, err
	}

	return num, nil
}

// ExistsInSlice returns whether a given string exists in the given slice
func ExistsInSlice(str string, sl []string) bool {
	for _, item := range sl {
		if item == str {
			return true
		}
	}
	return false
}
