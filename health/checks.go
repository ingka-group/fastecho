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

package health

import "gorm.io/gorm"

// checkDatabase pings the database and returns an error if it occurs.
// If a database doesn't exist, the function returns no error.
func checkDatabase(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDb, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDb.Ping()
}
