package test

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	bqMountPath = "/mnt/data.yaml"
	bqHttpPort  = "9050/tcp"
	bqGrpcPort  = "9060/tcp"
	bqProject   = "test"
)

type BigqueryEmulatorContainer struct {
	testcontainers.Container

	BqHost     string
	BqRestPort int
	BqGrpcPort int
}

func setupBigqueryEmulator(ctx context.Context, dataPath string) (*BigqueryEmulatorContainer, error) {
	executionPath, err := testpath()
	if err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{
		Image: "ghcr.io/goccy/bigquery-emulator:latest",
		HostConfigModifier: func(config *container.HostConfig) {
			config.Mounts = append(config.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: fmt.Sprintf("%s/%s", executionPath, dataPath),
				Target: bqMountPath,
			})
		},
		Cmd: []string{
			fmt.Sprintf("--project=%s", bqProject),
			fmt.Sprintf("--data-from-yaml=%s", bqMountPath),
		},
		ExposedPorts: []string{bqHttpPort, bqGrpcPort},
		WaitingFor:   wait.ForListeningPort(bqGrpcPort),
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

	mappedHttpPort, err := container.MappedPort(ctx, bqHttpPort)
	if err != nil {
		return nil, err
	}

	mappedGrpcPort, err := container.MappedPort(ctx, bqGrpcPort)
	if err != nil {
		return nil, err
	}

	return &BigqueryEmulatorContainer{
		Container:  container,
		BqHost:     hostIP,
		BqRestPort: mappedHttpPort.Int(),
		BqGrpcPort: mappedGrpcPort.Int(),
	}, nil
}
