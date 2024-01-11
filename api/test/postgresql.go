package test

import (
	"context"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ingka-group-digital/ocp-go-utils/stringutils"
)

const (
	dbName     = "postgres"
	dbUsername = "postgres"
	dbPassword = "password"
	dbPort     = "5432/tcp"
)

// PostgresDBContainer holds all the necessary information for postgres database test container.
type PostgresDBContainer struct {
	testcontainers.Container

	DBHost     string
	DBPort     int
	DBName     string
	DBUsername string
	DBPassword string
}

// setupPostgresDB sets up a postgres database test container.
func setupPostgresDB(ctx context.Context, initSQLScript ...string) (*PostgresDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "postgres:latest",
		Env: map[string]string{
			"POSTGRES_USER":     dbUsername,
			"POSTGRES_PASSWORD": dbPassword,
		},
		ExposedPorts: []string{dbPort},
		WaitingFor:   wait.ForSQL(dbPort, "postgres", dbURL),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	// If init script path is provided, initialize the database using the script.
	if !stringutils.IsEmpty(initSQLScript[0]) {
		err = initDB(container, initSQLScript[0])
		if err != nil {
			return nil, err
		}
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, dbPort)
	if err != nil {
		return nil, err
	}

	return &PostgresDBContainer{
		Container:  container,
		DBHost:     hostIP,
		DBPort:     mappedPort.Int(),
		DBName:     dbName,
		DBUsername: dbUsername,
		DBPassword: dbPassword,
	}, nil
}

// dbURL returns the postgres database URL.
func dbURL(host string, port nat.Port) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, host, port.Port(), dbName,
	)
}

// initDB initializes the database using the provided script.
func initDB(container testcontainers.Container, filename string) error {
	executionPath, err := testpath()
	if err != nil {
		return err
	}

	containerPath := fmt.Sprintf("/%s", filename)

	// Copy the script from the host path to the container
	err = container.CopyFileToContainer(
		context.Background(),
		fmt.Sprintf(
			"%s/fixtures/%s", executionPath, filename,
		),
		containerPath,
		0544,
	)
	if err != nil {
		return err
	}

	// Execute the script
	stdout, stderr, err := container.Exec(context.Background(), []string{
		"bash",
		"-c",
		fmt.Sprintf(
			"export PGPASSWORD=%s && psql -U %s -d %s -f %s",
			dbPassword, dbUsername, dbName, containerPath,
		),
	})

	log.Println("container exec [stdout]: ", stdout)
	log.Println("container exec [stderr]: ", stderr)

	if err != nil {
		return err
	}

	return err
}
