package test

import (
	"context"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName     = "postgres"
	dbUsername = "postgres"
	dbPassword = "password"
	dbPort     = "5432/tcp"
)

type PostgresDBContainer struct {
	testcontainers.Container

	DBHost     string
	DBPort     int
	DBName     string
	DBUsername string
	DBPassword string
}

func setupPostgresDB(ctx context.Context) (*PostgresDBContainer, error) {
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

func dbURL(host string, port nat.Port) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, host, port.Port(), dbName,
	)
}
