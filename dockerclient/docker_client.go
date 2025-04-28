package dockerclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var defaultClient *client.Client

func DefaultClient() (*client.Client, error) {
	if defaultClient == nil {
		c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("failed to create docker client: %w", err)
		}
		defaultClient = c
	}

	return defaultClient, nil
}

var ErrNoContainerFound = fmt.Errorf("no container found")

func GetContainer(ctx context.Context, containerName string) (container.Summary, error) {
	dockerClient, err := DefaultClient()
	if err != nil {
		return container.Summary{}, fmt.Errorf("failed to get docker client: %w", err)
	}
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{Filters: filters.NewArgs(
		filters.Arg("name", containerName),
	)})
	if err != nil {
		return container.Summary{}, fmt.Errorf("failed to list containers: %w", err)
	}
	for _, c := range containers {
		if slices.Contains(c.Names, "/"+containerName) {
			return c, nil
		}
	}
	return container.Summary{}, ErrNoContainerFound
}

func ListAllContainers(ctx context.Context) ([]container.Summary, error) {
	dockerClient, err := DefaultClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get docker client: %w", err)
	}
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	return containers, nil
}

type ContainerStatus struct {
	Running bool
}

func GetStatus(ctx context.Context, containerName string) (*container.InspectResponse, error) {
	dockerClient, err := DefaultClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get docker client: %w", err)
	}
	c, err := GetContainer(ctx, containerName)
	if err != nil {
		if errors.Is(err, ErrNoContainerFound) {
			slog.With(slog.String("container", containerName)).Debug("No container found")
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}
	inspection, err := dockerClient.ContainerInspect(context.Background(), c.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	return &inspection, nil
}

func StopContainer(ctx context.Context, containerName string) error {
	dockerClient, err := DefaultClient()
	if err != nil {
		return fmt.Errorf("failed to get docker client: %w", err)
	}
	c, err := GetContainer(ctx, containerName)
	if err != nil {
		if errors.Is(err, ErrNoContainerFound) {
			slog.With(slog.String("container", containerName)).Info("No container found")
			return nil
		}
		return fmt.Errorf("failed to get container: %w", err)
	}
	err = dockerClient.ContainerKill(context.Background(), c.ID, "")
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}
