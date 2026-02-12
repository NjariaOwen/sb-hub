package pkg

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	archive "github.com/moby/go-archive"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerClient interface {
	Ping(ctx context.Context) (types.Ping, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	NetworkInspect(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error)
	NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
}

type Dockerengine struct {
	Client DockerClient
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

	io.Copy(os.Stdout, out)
	return nil
}

func (e *Dockerengine) RemoveSandbox(ctx context.Context, name string, storagePath string, forceAll bool) error {
	stopTimeout := 30
	e.Client.ContainerStop(ctx, name, container.StopOptions{Timeout: &stopTimeout})
	err := e.Client.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}

	if forceAll && storagePath != "" {
		return os.RemoveAll(storagePath)
	}
	return nil
}

func (e *Dockerengine) CreateSandbox(ctx context.Context, name string, ttl time.Duration, size string, config *container.Config, hostConfig *container.HostConfig) (string, error) {
	expiry := time.Now().Add(ttl).Format(time.RFC3339)
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	config.Labels["com.sbhub.managed"] = "true"
	config.Labels["com.sbhub.expires"] = expiry
	config.Labels["com.sbhub.size"] = size
	config.Tty = true
	config.OpenStdin = true

	resp, err := e.Client.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", err
	}

	err = e.Client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	return resp.ID, err
}

func (e *Dockerengine) EnsureNetwork(ctx context.Context) error {
	netName := "sb-hub-net"
	_, err := e.Client.NetworkInspect(ctx, netName, network.InspectOptions{})
	if err != nil {
		_, err = e.Client.NetworkCreate(ctx, netName, network.CreateOptions{
			Driver: "bridge",
			Labels: map[string]string{"com.sbhub.managed": "true"},
		})
	}
	return err
}

func (e *Dockerengine) BuildImage(ctx context.Context, path, tag string) error {
	fmt.Printf("🛠️  Building custom image: %s\n", tag)
	tar, err := archive.TarWithOptions(path, &archive.TarOptions{})
	if err != nil {
		return err
	}
	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
		Remove:     true,
	}
	res, err := e.Client.ImageBuild(ctx, tar, opts)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
	return nil
}

func (e *Dockerengine) GetUsedPorts(ctx context.Context) (map[string]bool, error) {
	containers, err := e.Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	used := make(map[string]bool)
	for _, c := range containers {
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				used[fmt.Sprintf("%d", p.PublicPort)] = true
			}
		}
	}
	return used, nil
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

func (e *Dockerengine) GetExpiredSandboxes(ctx context.Context) ([]container.Summary, error) {
	f := filters.NewArgs()
	f.Add("label", "com.sbhub.managed=true")
	containers, err := e.Client.ContainerList(ctx, container.ListOptions{Filters: f})
	if err != nil {
		return nil, err
	}

	var expired []container.Summary
	now := time.Now()
	for _, c := range containers {
		if expStr, ok := c.Labels["com.sbhub.expires"]; ok {
			expiry, err := time.Parse(time.RFC3339, expStr)
			if err == nil && now.After(expiry) {
				expired = append(expired, c)
			}
		}
	}
	return expired, nil
}

func (e *Dockerengine) InspectSandbox(ctx context.Context, name string) (container.InspectResponse, error) {
	return e.Client.ContainerInspect(ctx, name)
}

func GenerateRandomName() string {
	adjectives := []string{"swift", "brave", "cool", "mighty", "keen"}
	nouns := []string{"whale", "ship", "anchor", "pilot", "wave"}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s-%s-%d", adjectives[r.Intn(len(adjectives))], nouns[r.Intn(len(nouns))], r.Intn(1000))
}
