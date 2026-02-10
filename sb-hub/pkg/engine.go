package pkg

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type Dockerengine struct {
	Client *client.Client
}

func (e *Dockerengine) Ping(ctx context.Context) error {
	_, err := e.Client.Ping(ctx)
	return err
}

func (e *Dockerengine) ContainerExists(ctx context.Context, name string) (bool, error) {
	filter := filters.NewArgs()
	filter.Add("name", fmt.Sprintf("^/%s$", name))

	containers, err := e.Client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return false, err
	}
	return len(containers) > 0, nil
}

func (e *Dockerengine) EnsureImage(ctx context.Context, imageName string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	fmt.Printf("🔍 Checking for image '%s'...\n", imageName)

	out, err := e.Client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(os.Stdout, out)
	return err
}

func (e *Dockerengine) RemoveSandbox(ctx context.Context, name string, storagePath string, forceAll bool) error {
	stopTimeout := 30
	fmt.Printf("🛑 Stopping %s (30s timeout)...\n", name)
	e.Client.ContainerStop(ctx, name, container.StopOptions{Timeout: &stopTimeout})

	fmt.Printf("🗑️  Removing container %s...\n", name)
	err := e.Client.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}

	if forceAll {
		fmt.Printf("🧹 Wiping persistent data at %s...\n", storagePath)
		return os.RemoveAll(storagePath)
	}
	return nil
}

func (e *Dockerengine) CreateSandbox(ctx context.Context, name string, ttl time.Duration, config *container.Config, hostConfig *container.HostConfig) (string, error) {
	expiry := time.Now().Add(ttl).Format(time.RFC3339)
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	config.Labels["com.sbhub.managed"] = "true"
	config.Labels["com.sbhub.expires"] = expiry

	resp, err := e.Client.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", err
	}

	err = e.Client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	return resp.ID, err
}

func (e *Dockerengine) GetActiveSandboxes(ctx context.Context) (map[string]container.Summary, error) {
	containers, err := e.Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	active := make(map[string]container.Summary)
	for _, c := range containers {
		for _, name := range c.Names {
			active[filepath.Base(name)] = c
		}
	}
	return active, nil
}
