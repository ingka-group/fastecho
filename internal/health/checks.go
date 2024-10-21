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
