package health

import "gorm.io/gorm"

// CheckDatabase pings the database and returns an error if it occurs.
// If the database is nil, it returns no error.
func CheckDatabase(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDb, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDb.Ping()
}
