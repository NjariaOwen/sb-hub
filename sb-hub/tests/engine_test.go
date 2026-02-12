package tests

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ---------------------------------------------------------------------------
// Mock Docker Client
// ---------------------------------------------------------------------------

type MockDockerClient struct {
	PingFn             func(ctx context.Context) (types.Ping, error)
	ContainerListFn    func(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerCreateFn  func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStartFn   func(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStopFn    func(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemoveFn  func(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerInspectFn func(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerLogsFn    func(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	ImagePullFn        func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ImageBuildFn       func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	NetworkInspectFn   func(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error)
	NetworkCreateFn    func(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
}

func (m *MockDockerClient) Ping(ctx context.Context) (types.Ping, error) {
	if m.PingFn != nil {
		return m.PingFn(ctx)
	}
	return types.Ping{}, nil
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	if m.ContainerListFn != nil {
		return m.ContainerListFn(ctx, options)
	}
	return nil, nil
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	if m.ContainerCreateFn != nil {
		return m.ContainerCreateFn(ctx, config, hostConfig, networkingConfig, platform, containerName)
	}
	return container.CreateResponse{ID: "mock-id-123456789012"}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	if m.ContainerStartFn != nil {
		return m.ContainerStartFn(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	if m.ContainerStopFn != nil {
		return m.ContainerStopFn(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if m.ContainerRemoveFn != nil {
		return m.ContainerRemoveFn(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	if m.ContainerInspectFn != nil {
		return m.ContainerInspectFn(ctx, containerID)
	}
	return container.InspectResponse{}, nil
}

func (m *MockDockerClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	if m.ContainerLogsFn != nil {
		return m.ContainerLogsFn(ctx, containerID, options)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	if m.ImagePullFn != nil {
		return m.ImagePullFn(ctx, refStr, options)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	if m.ImageBuildFn != nil {
		return m.ImageBuildFn(ctx, buildContext, options)
	}
	return types.ImageBuildResponse{Body: io.NopCloser(strings.NewReader(""))}, nil
}

func (m *MockDockerClient) NetworkInspect(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error) {
	if m.NetworkInspectFn != nil {
		return m.NetworkInspectFn(ctx, networkID, options)
	}
	return network.Inspect{}, nil
}

func (m *MockDockerClient) NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
	if m.NetworkCreateFn != nil {
		return m.NetworkCreateFn(ctx, name, options)
	}
	return network.CreateResponse{}, nil
}

// ---------------------------------------------------------------------------
// Tests: Ping
// ---------------------------------------------------------------------------

func TestPing_Success(t *testing.T) {
	mock := &MockDockerClient{
		PingFn: func(ctx context.Context) (types.Ping, error) {
			return types.Ping{}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.Ping(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPing_Error(t *testing.T) {
	mock := &MockDockerClient{
		PingFn: func(ctx context.Context) (types.Ping, error) {
			return types.Ping{}, errors.New("connection refused")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "connection refused" {
		t.Fatalf("expected 'connection refused', got '%v'", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: ContainerExists
// ---------------------------------------------------------------------------

func TestContainerExists_Found(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{{ID: "abc123"}}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	exists, err := engine.ContainerExists(context.Background(), "my-sandbox")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected container to exist")
	}
}

func TestContainerExists_NotFound(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	exists, err := engine.ContainerExists(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected container to not exist")
	}
}

func TestContainerExists_Error(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return nil, errors.New("docker daemon down")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	_, err := engine.ContainerExists(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: EnsureImage
// ---------------------------------------------------------------------------

func TestEnsureImage_Success(t *testing.T) {
	pullCalled := false
	mock := &MockDockerClient{
		ImagePullFn: func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
			pullCalled = true
			if refStr != "alpine:latest" {
				t.Fatalf("expected image 'alpine:latest', got '%s'", refStr)
			}
			return io.NopCloser(strings.NewReader("pulling...")), nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.EnsureImage(context.Background(), "alpine:latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pullCalled {
		t.Fatal("expected ImagePull to be called")
	}
}

func TestEnsureImage_Error(t *testing.T) {
	mock := &MockDockerClient{
		ImagePullFn: func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
			return nil, errors.New("image not found")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.EnsureImage(context.Background(), "nonexistent:latest")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: RemoveSandbox
// ---------------------------------------------------------------------------

func TestRemoveSandbox_Basic(t *testing.T) {
	stopCalled := false
	removeCalled := false
	mock := &MockDockerClient{
		ContainerStopFn: func(ctx context.Context, containerID string, options container.StopOptions) error {
			stopCalled = true
			return nil
		},
		ContainerRemoveFn: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			removeCalled = true
			if containerID != "my-sandbox" {
				t.Fatalf("expected container name 'my-sandbox', got '%s'", containerID)
			}
			return nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.RemoveSandbox(context.Background(), "my-sandbox", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stopCalled {
		t.Fatal("expected ContainerStop to be called")
	}
	if !removeCalled {
		t.Fatal("expected ContainerRemove to be called")
	}
}

func TestRemoveSandbox_ForceAll(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/testfile"
	os.WriteFile(testFile, []byte("data"), 0644)

	mock := &MockDockerClient{}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.RemoveSandbox(context.Background(), "test", tmpDir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Fatal("expected storage directory to be deleted")
	}
}

func TestRemoveSandbox_RemoveError(t *testing.T) {
	mock := &MockDockerClient{
		ContainerRemoveFn: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			return errors.New("container busy")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.RemoveSandbox(context.Background(), "test", "", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: CreateSandbox
// ---------------------------------------------------------------------------

func TestCreateSandbox_Success(t *testing.T) {
	expectedID := "abc123def456ghi789"
	mock := &MockDockerClient{
		ContainerCreateFn: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
			if containerName != "my-sandbox" {
				t.Fatalf("expected name 'my-sandbox', got '%s'", containerName)
			}
			if config.Labels["com.sbhub.managed"] != "true" {
				t.Fatal("expected managed label to be set")
			}
			if config.Labels["com.sbhub.size"] != "small" {
				t.Fatalf("expected size label 'small', got '%s'", config.Labels["com.sbhub.size"])
			}
			if _, ok := config.Labels["com.sbhub.expires"]; !ok {
				t.Fatal("expected expires label to be set")
			}
			if !config.Tty {
				t.Fatal("expected Tty to be true")
			}
			if !config.OpenStdin {
				t.Fatal("expected OpenStdin to be true")
			}
			return container.CreateResponse{ID: expectedID}, nil
		},
		ContainerStartFn: func(ctx context.Context, containerID string, options container.StartOptions) error {
			if containerID != expectedID {
				t.Fatalf("expected start with ID '%s', got '%s'", expectedID, containerID)
			}
			return nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	config := &container.Config{
		Image: "alpine:latest",
	}
	hostConfig := &container.HostConfig{}

	id, err := engine.CreateSandbox(context.Background(), "my-sandbox", 1*time.Hour, "small", config, hostConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Fatalf("expected ID '%s', got '%s'", expectedID, id)
	}
}

func TestCreateSandbox_ExpiryLabel(t *testing.T) {
	beforeCreate := time.Now()

	mock := &MockDockerClient{
		ContainerCreateFn: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
			expiryStr := config.Labels["com.sbhub.expires"]
			expiry, err := time.Parse(time.RFC3339, expiryStr)
			if err != nil {
				t.Fatalf("failed to parse expiry label: %v", err)
			}
			expectedExpiry := beforeCreate.Add(2 * time.Hour)
			if expiry.Before(expectedExpiry.Add(-5 * time.Second)) {
				t.Fatalf("expiry %v is too early, expected around %v", expiry, expectedExpiry)
			}
			return container.CreateResponse{ID: "test-id-000000000000"}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	config := &container.Config{Image: "alpine:latest"}
	hostConfig := &container.HostConfig{}

	_, err := engine.CreateSandbox(context.Background(), "test", 2*time.Hour, "medium", config, hostConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateSandbox_CreateError(t *testing.T) {
	mock := &MockDockerClient{
		ContainerCreateFn: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
			return container.CreateResponse{}, errors.New("name conflict")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	_, err := engine.CreateSandbox(context.Background(), "test", 1*time.Hour, "small", &container.Config{Image: "alpine"}, &container.HostConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSandbox_StartError(t *testing.T) {
	mock := &MockDockerClient{
		ContainerCreateFn: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
			return container.CreateResponse{ID: "test-id-000000000000"}, nil
		},
		ContainerStartFn: func(ctx context.Context, containerID string, options container.StartOptions) error {
			return errors.New("port already in use")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	_, err := engine.CreateSandbox(context.Background(), "test", 1*time.Hour, "small", &container.Config{Image: "alpine"}, &container.HostConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: EnsureNetwork
// ---------------------------------------------------------------------------

func TestEnsureNetwork_Exists(t *testing.T) {
	createCalled := false
	mock := &MockDockerClient{
		NetworkInspectFn: func(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error) {
			return network.Inspect{Name: "sb-hub-net"}, nil
		},
		NetworkCreateFn: func(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
			createCalled = true
			return network.CreateResponse{}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.EnsureNetwork(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createCalled {
		t.Fatal("NetworkCreate should not be called when network already exists")
	}
}

func TestEnsureNetwork_Creates(t *testing.T) {
	createCalled := false
	mock := &MockDockerClient{
		NetworkInspectFn: func(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error) {
			return network.Inspect{}, errors.New("network not found")
		},
		NetworkCreateFn: func(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
			createCalled = true
			if name != "sb-hub-net" {
				t.Fatalf("expected network name 'sb-hub-net', got '%s'", name)
			}
			if options.Driver != "bridge" {
				t.Fatalf("expected bridge driver, got '%s'", options.Driver)
			}
			if options.Labels["com.sbhub.managed"] != "true" {
				t.Fatal("expected managed label on network")
			}
			return network.CreateResponse{}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.EnsureNetwork(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !createCalled {
		t.Fatal("expected NetworkCreate to be called")
	}
}

// ---------------------------------------------------------------------------
// Tests: GetUsedPorts
// ---------------------------------------------------------------------------

func TestGetUsedPorts(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{
				{
					Ports: []container.Port{
						{PublicPort: 8080},
						{PublicPort: 8081},
					},
				},
				{
					Ports: []container.Port{
						{PublicPort: 0},
						{PublicPort: 9090},
					},
				},
			}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	ports, err := engine.GetUsedPorts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]bool{"8080": true, "8081": true, "9090": true}
	if len(ports) != len(expected) {
		t.Fatalf("expected %d ports, got %d", len(expected), len(ports))
	}
	for k := range expected {
		if !ports[k] {
			t.Fatalf("expected port %s to be marked used", k)
		}
	}
	if ports["0"] {
		t.Fatal("port 0 should not be in the used map")
	}
}

func TestGetUsedPorts_Error(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return nil, errors.New("daemon unreachable")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	_, err := engine.GetUsedPorts(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: GetActiveSandboxes
// ---------------------------------------------------------------------------

func TestGetActiveSandboxes(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{
				{
					ID:    "id1",
					Names: []string{"/sandbox-a"},
					State: "running",
				},
				{
					ID:    "id2",
					Names: []string{"/sandbox-b"},
					State: "exited",
				},
			}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	active, err := engine.GetActiveSandboxes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 sandboxes, got %d", len(active))
	}
	if _, ok := active["sandbox-a"]; !ok {
		t.Fatal("expected 'sandbox-a' in results")
	}
	if _, ok := active["sandbox-b"]; !ok {
		t.Fatal("expected 'sandbox-b' in results")
	}
}

// ---------------------------------------------------------------------------
// Tests: GetExpiredSandboxes
// ---------------------------------------------------------------------------

func TestGetExpiredSandboxes_Mixed(t *testing.T) {
	pastExpiry := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	futureExpiry := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{
				{
					ID:     "expired-1",
					Names:  []string{"/old-sandbox"},
					Labels: map[string]string{"com.sbhub.managed": "true", "com.sbhub.expires": pastExpiry},
				},
				{
					ID:     "valid-1",
					Names:  []string{"/new-sandbox"},
					Labels: map[string]string{"com.sbhub.managed": "true", "com.sbhub.expires": futureExpiry},
				},
			}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	expired, err := engine.GetExpiredSandboxes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired sandbox, got %d", len(expired))
	}
	if expired[0].ID != "expired-1" {
		t.Fatalf("expected expired sandbox 'expired-1', got '%s'", expired[0].ID)
	}
}

func TestGetExpiredSandboxes_NoneExpired(t *testing.T) {
	futureExpiry := time.Now().Add(5 * time.Hour).Format(time.RFC3339)

	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{
				{
					ID:     "valid-1",
					Labels: map[string]string{"com.sbhub.managed": "true", "com.sbhub.expires": futureExpiry},
				},
			}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	expired, err := engine.GetExpiredSandboxes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expired) != 0 {
		t.Fatalf("expected 0 expired sandboxes, got %d", len(expired))
	}
}

func TestGetExpiredSandboxes_InvalidExpiry(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{
				{
					ID:     "bad-1",
					Labels: map[string]string{"com.sbhub.managed": "true", "com.sbhub.expires": "not-a-date"},
				},
			}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	expired, err := engine.GetExpiredSandboxes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expired) != 0 {
		t.Fatalf("expected 0 expired (invalid date should be skipped), got %d", len(expired))
	}
}

func TestGetExpiredSandboxes_NoExpiryLabel(t *testing.T) {
	mock := &MockDockerClient{
		ContainerListFn: func(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{
				{
					ID:     "no-expiry",
					Labels: map[string]string{"com.sbhub.managed": "true"},
				},
			}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	expired, err := engine.GetExpiredSandboxes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expired) != 0 {
		t.Fatalf("expected 0 expired (no expiry label), got %d", len(expired))
	}
}

// ---------------------------------------------------------------------------
// Tests: InspectSandbox
// ---------------------------------------------------------------------------

func TestInspectSandbox_Success(t *testing.T) {
	mock := &MockDockerClient{
		ContainerInspectFn: func(ctx context.Context, containerID string) (container.InspectResponse, error) {
			return container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					ID:   "abc123",
					Name: "/my-sandbox",
				},
			}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	resp, err := engine.InspectSandbox(context.Background(), "my-sandbox")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "abc123" {
		t.Fatalf("expected ID 'abc123', got '%s'", resp.ID)
	}
}

func TestInspectSandbox_NotFound(t *testing.T) {
	mock := &MockDockerClient{
		ContainerInspectFn: func(ctx context.Context, containerID string) (container.InspectResponse, error) {
			return container.InspectResponse{}, errors.New("No such container: nonexistent")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	_, err := engine.InspectSandbox(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: GenerateRandomName
// ---------------------------------------------------------------------------

func TestGenerateRandomName_Format(t *testing.T) {
	pattern := regexp.MustCompile(`^(swift|brave|cool|mighty|keen)-(whale|ship|anchor|pilot|wave)-\d+$`)

	for i := 0; i < 50; i++ {
		name := pkg.GenerateRandomName()
		if !pattern.MatchString(name) {
			t.Fatalf("name '%s' does not match expected pattern", name)
		}
	}
}

func TestGenerateRandomName_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := pkg.GenerateRandomName()
		seen[name] = true
	}
	if len(seen) < 2 {
		t.Fatal("expected at least some variety in generated names")
	}
}

// ---------------------------------------------------------------------------
// Tests: BuildImage
// ---------------------------------------------------------------------------

func TestBuildImage_Success(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(fmt.Sprintf("%s/Dockerfile", tmpDir), []byte("FROM alpine:latest\n"), 0644)

	buildCalled := false
	mock := &MockDockerClient{
		ImageBuildFn: func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
			buildCalled = true
			if len(options.Tags) != 1 || options.Tags[0] != "my-tag" {
				t.Fatalf("expected tag 'my-tag', got %v", options.Tags)
			}
			if options.Dockerfile != "Dockerfile" {
				t.Fatalf("expected Dockerfile, got '%s'", options.Dockerfile)
			}
			return types.ImageBuildResponse{Body: io.NopCloser(strings.NewReader("build output"))}, nil
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.BuildImage(context.Background(), tmpDir, "my-tag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !buildCalled {
		t.Fatal("expected ImageBuild to be called")
	}
}

func TestBuildImage_BuildError(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(fmt.Sprintf("%s/Dockerfile", tmpDir), []byte("FROM alpine:latest\n"), 0644)

	mock := &MockDockerClient{
		ImageBuildFn: func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
			return types.ImageBuildResponse{}, errors.New("build failed")
		},
	}
	engine := &pkg.Dockerengine{Client: mock}

	err := engine.BuildImage(context.Background(), tmpDir, "bad-tag")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
