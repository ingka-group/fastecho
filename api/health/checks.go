package health

import "gorm.io/gorm"

// CheckDatabase pings the database and returns an error if it occurs.
func CheckDatabase(db *gorm.DB) error {
	sqlDb, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDb.Ping()
}
