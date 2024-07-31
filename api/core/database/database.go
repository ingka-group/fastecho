package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ingka-group-digital/ocp-go-utils/api/core/env"
)

const (
	dbName            = "DB_NAME"
	dbHostname        = "DB_HOST"
	dbPort            = "DB_PORT"
	dbSSLMode         = "DB_SSL_MODE"
	dbUsername        = "DB_READ_WRITE_USER"
	dbPassword        = "DB_READ_WRITE_PASSWORD"
	dbMaxOpenConn     = "DB_MAX_OPEN_CONNECTIONS"
	dbMaxIdleConn     = "DB_MAX_IDLE_CONNECTIONS"
	dbMaxConnLifeTime = "DB_CONNECTION_MAX_LIFETIME"
)

var (
	dbEnv = env.EnvVars{
		dbHostname: {
			DefaultValue: "localhost",
		},
		dbPort: {
			DefaultValue: "5432",
			IsInteger:    true,
		},
		dbName:     {},
		dbUsername: {},
		dbPassword: {},
		dbSSLMode: {
			DefaultValue: "disable",
			OneOf:        []string{"enable", "disable"},
		},
		dbMaxOpenConn: {
			DefaultValue: "10",
			IsInteger:    true,
		},
		dbMaxIdleConn: {
			DefaultValue: "10",
			IsInteger:    true,
		},
		dbMaxConnLifeTime: {
			DefaultValue: "1h",
		},
	}
)

// DBConfig is the struct that contains the database configuration.
type dbConfig struct {
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

// NewDB creates a database configuration from the environment variables.
func NewDB() (*gorm.DB, error) {
	var db *gorm.DB

	// options are not used here
	env, err := env.SetEnv(dbEnv)
	if err != nil {
		return nil, err
	}

	lifetime, err := time.ParseDuration(env[dbMaxConnLifeTime].Value)
	if err != nil {
		lifetime = time.Hour
	}

	dbConfig := &dbConfig{
		Hostname:        env[dbHostname].Value,
		Port:            env[dbPort].IntValue,
		Name:            env[dbName].Value,
		Username:        env[dbUsername].Value,
		Password:        env[dbPassword].Value,
		SSLMode:         env[dbSSLMode].Value,
		TimeZone:        time.UTC,
		MaxIdleConn:     env[dbMaxIdleConn].IntValue,
		MaxOpenedConn:   env[dbMaxOpenConn].IntValue,
		ConnMaxLifetime: lifetime,
	}

	db, err = dbConfig.setup()
	if err != nil {
		return nil, err
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}

	err = migrateDB(sqlDb)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// BuildDSN builds the Data Source Name (DSN) which represents the database connection string.
func (c *dbConfig) buildDSN() (string, error) {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%v TimeZone=%s",
		c.Hostname,
		c.Username,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode,
		c.TimeZone), nil
}

// setup creates a new database based on the configuration given.
func (c *dbConfig) setup() (*gorm.DB, error) {
	dsn, err := c.buildDSN()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDb.SetMaxIdleConns(c.MaxIdleConn)
	sqlDb.SetMaxOpenConns(c.MaxOpenedConn)
	sqlDb.SetConnMaxLifetime(c.ConnMaxLifetime)

	return db, nil
}
