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

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ingka-group/fastecho/env"
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
	dbEnvs = env.Map{
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

// NewDB creates a new *gorm.DB the configuration of which is through environment variables.
func NewDB(cfg *gorm.Config) (*gorm.DB, error) {
	var db *gorm.DB

	// options are not used here
	err := dbEnvs.SetEnv()
	if err != nil {
		return nil, err
	}

	lifetime, err := time.ParseDuration(dbEnvs[dbMaxConnLifeTime].Value)
	if err != nil {
		lifetime = time.Hour
	}

	dbConf := &dbConfig{
		Hostname:        dbEnvs[dbHostname].Value,
		Port:            dbEnvs[dbPort].IntValue,
		Name:            dbEnvs[dbName].Value,
		Username:        dbEnvs[dbUsername].Value,
		Password:        dbEnvs[dbPassword].Value,
		SSLMode:         dbEnvs[dbSSLMode].Value,
		TimeZone:        time.UTC,
		MaxIdleConn:     dbEnvs[dbMaxIdleConn].IntValue,
		MaxOpenedConn:   dbEnvs[dbMaxOpenConn].IntValue,
		ConnMaxLifetime: lifetime,
	}

	db, err = dbConf.setup(cfg)
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

// dbConfig contains the database configuration.
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

// setup creates a new database based on the configuration given.
func (c *dbConfig) setup(cfg *gorm.Config) (*gorm.DB, error) {
	dsn, err := c.buildDSN()
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}

	db, err := gorm.Open(
		postgres.Open(dsn),
		cfg,
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

// migrateDB migrates the database to the latest version using goose.
func migrateDB(db *sql.DB) error {
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		os.DirFS("db/migrations"),
	)

	if err != nil {
		if errors.Is(err, goose.ErrNoMigrations) {
			return nil
		}
		return err
	}

	results, err := provider.Up(context.Background())
	if err != nil {
		return err
	}

	for i := range results {
		if results[i].Error != nil {
			return results[i].Error
		}

		log.Println("[", i+1, "] Migration applied:", results[i].Source.Path)
	}

	return nil
}
