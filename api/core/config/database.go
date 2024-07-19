package config

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig is the struct that contains the database configuration.
type DBConfig struct {
	Hostname        string
	Port            int
	Name            string
	Username        string
	Password        string
	SSLMode         string
	TimeZone        *time.Location
	MaxIdleConn     int
	MaxOpenedConn   int
	ConnMaxLifetime time.Duration
}

// NewDBConfig creates a database configuration from the environment variables.
func NewDBConfig() (*DBConfig, error) {
	lifetime, err := time.ParseDuration(Env[DBMaxConnLifeTime].Value)
	if err != nil {
		lifetime = time.Hour
	}

	return &DBConfig{
		Hostname:        Env[DBHostname].Value,
		Port:            Env[DBPort].IntValue,
		Name:            Env[DBName].Value,
		Username:        Env[DBUsername].Value,
		Password:        Env[DBPassword].Value,
		SSLMode:         Env[DBSSLMode].Value,
		TimeZone:        time.UTC,
		MaxIdleConn:     Env[DBMaxIdleConn].IntValue,
		MaxOpenedConn:   Env[DBMaxOpenConn].IntValue,
		ConnMaxLifetime: lifetime,
	}, nil
}

// BuildDSN builds the Data Source Name (DSN) which represents the database connection string.
func (c *DBConfig) BuildDSN() (string, error) {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%v TimeZone=%s",
		c.Hostname,
		c.Username,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode,
		c.TimeZone), nil
}

// NewDB creates a new database based on the configuration given.
func NewDB(DbConfig *DBConfig) (*gorm.DB, error) {
	dsn, err := DbConfig.BuildDSN()
	if err != nil {
		return nil, err
	}

	DB, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDb, err := DB.DB()
	if err != nil {
		return nil, err
	}

	sqlDb.SetMaxIdleConns(DbConfig.MaxIdleConn)
	sqlDb.SetMaxOpenConns(DbConfig.MaxOpenedConn)
	sqlDb.SetConnMaxLifetime(DbConfig.ConnMaxLifetime)

	return DB, nil
}
