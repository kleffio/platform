// Package docker implements RuntimeAdapter using the Docker Engine API.
// Requires /var/run/docker.sock to be mounted into the platform container.
package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kleffio/platform/internal/shared/runtime"
)

// Runtime implements runtime.RuntimeAdapter using the Docker Engine API.
type Runtime struct {
	client      *dockerclient.Client
	networkName string
	logger      *slog.Logger
}

// New creates a DockerRuntime. networkName is the Docker network plugins are
// attached to (default: "kleff").
func New(networkName string, logger *slog.Logger) (*Runtime, error) {
	if networkName == "" {
		networkName = "kleff"
	}
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("docker: create client: %w", err)
	}
	return &Runtime{client: cli, networkName: networkName, logger: logger}, nil
}

var _ runtime.RuntimeAdapter = (*Runtime)(nil)

// Deploy pulls the image and starts the container, replacing any existing one.
func (r *Runtime) Deploy(ctx context.Context, spec runtime.ContainerSpec) error {
	// Remove any existing container with this ID first (idempotent).
	_ = r.Remove(ctx, spec.ID)

	// Pull image (attempt; fall back to cached if pull fails).
	if err := r.pullImage(ctx, spec.Image); err != nil {
		r.logger.Warn("docker: image pull failed, using cached image if available",
			"image", spec.Image, "error", err)
	}

	// Build env slice.
	env := make([]string, 0, len(spec.Env))
	for k, v := range spec.Env {
		env = append(env, k+"="+v)
	}

	// Build port bindings.
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for _, p := range spec.Ports {
		key := nat.Port(fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))
		exposedPorts[key] = struct{}{}
		if p.HostPort != 0 {
			portBindings[key] = []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", p.HostPort)},
			}
		}
	}

	// Build labels.
	labels := map[string]string{
		"kleff.io/managed":   "true",
		"kleff.io/plugin-id": spec.ID,
	}
	for k, v := range spec.Labels {
		labels[k] = v
	}

	restartPolicy := container.RestartPolicy{Name: container.RestartPolicyAlways}
	switch spec.RestartPolicy {
	case runtime.RestartOnFailure:
		restartPolicy = container.RestartPolicy{Name: container.RestartPolicyOnFailure}
	case runtime.RestartNever:
		restartPolicy = container.RestartPolicy{Name: container.RestartPolicyDisabled}
	}

	hostCfg := &container.HostConfig{
		PortBindings:  portBindings,
		RestartPolicy: restartPolicy,
	}
	if spec.Resources.MemoryMB > 0 {
		hostCfg.Memory = spec.Resources.MemoryMB * 1024 * 1024
	}
	if spec.Resources.CPUMillicores > 0 {
		hostCfg.NanoCPUs = spec.Resources.CPUMillicores * 1_000_000
	}

	resp, err := r.client.ContainerCreate(ctx,
		&container.Config{
			Image:        spec.Image,
			Env:          env,
			Labels:       labels,
			ExposedPorts: exposedPorts,
		},
		hostCfg,
		&network.NetworkingConfig{},
		nil,
		spec.ID,
	)
	if err != nil {
		return fmt.Errorf("docker: create container %q: %w", spec.ID, err)
	}

	// Attach to the kleff network.
	if err := r.client.NetworkConnect(ctx, r.networkName, resp.ID, nil); err != nil {
		// Non-fatal if the network doesn't exist yet; log and continue.
		r.logger.Warn("docker: network connect failed", "network", r.networkName,
			"container", spec.ID, "error", err)
	}

	if err := r.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("docker: start container %q: %w", spec.ID, err)
	}

	r.logger.Info("docker: container deployed", "id", spec.ID, "image", spec.Image)
	return nil
}

// Remove stops and removes the container. Returns nil if not found.
func (r *Runtime) Remove(ctx context.Context, id string) error {
	err := r.client.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
	if err != nil && !dockerclient.IsErrNotFound(err) {
		return fmt.Errorf("docker: remove container %q: %w", id, err)
	}
	return nil
}

// Start starts a stopped container.
func (r *Runtime) Start(ctx context.Context, id string) error {
	if err := r.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		if dockerclient.IsErrNotFound(err) {
			return fmt.Errorf("docker: container %q not found", id)
		}
		return fmt.Errorf("docker: start container %q: %w", id, err)
	}
	return nil
}

// Stop stops a running container gracefully.
func (r *Runtime) Stop(ctx context.Context, id string) error {
	timeout := 10
	if err := r.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout}); err != nil {
		if dockerclient.IsErrNotFound(err) {
			return nil // already gone
		}
		return fmt.Errorf("docker: stop container %q: %w", id, err)
	}
	return nil
}

// Status returns the current state of the container.
func (r *Runtime) Status(ctx context.Context, id string) (runtime.ContainerStatus, error) {
	info, err := r.client.ContainerInspect(ctx, id)
	if err != nil {
		if dockerclient.IsErrNotFound(err) {
			return runtime.ContainerStatus{ID: id, State: runtime.StateNotFound}, nil
		}
		return runtime.ContainerStatus{}, fmt.Errorf("docker: inspect container %q: %w", id, err)
	}

	var state runtime.ContainerState
	var since time.Time
	var msg string

	switch {
	case info.State.Running:
		state = runtime.StateRunning
		since, _ = time.Parse(time.RFC3339Nano, info.State.StartedAt)
	case info.State.Status == "created" || info.State.Status == "restarting":
		state = runtime.StateStarting
	case info.State.ExitCode != 0:
		state = runtime.StateFailed
		msg = fmt.Sprintf("exit code %d: %s", info.State.ExitCode, info.State.Error)
		since, _ = time.Parse(time.RFC3339Nano, info.State.FinishedAt)
	default:
		state = runtime.StateStopped
		since, _ = time.Parse(time.RFC3339Nano, info.State.FinishedAt)
	}

	return runtime.ContainerStatus{
		ID:      id,
		State:   state,
		Image:   info.Config.Image,
		Since:   since,
		Message: msg,
	}, nil
}

// Endpoint returns the Docker-network hostname:port for the container.
// Docker's embedded DNS resolves the container name within the shared network.
func (r *Runtime) Endpoint(_ context.Context, id string, port int) (string, error) {
	return fmt.Sprintf("%s:%d", id, port), nil
}

// Logs returns the last n lines of container stdout+stderr.
func (r *Runtime) Logs(ctx context.Context, id string, lines int) ([]string, error) {
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", lines),
	}
	rc, err := r.client.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, fmt.Errorf("docker: logs %q: %w", id, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("docker: read logs %q: %w", id, err)
	}

	raw := string(data)
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		// Strip 8-byte Docker multiplexing header.
		if len(line) > 8 {
			line = line[8:]
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out, nil
}

// pullImage pulls a Docker image, streaming progress to the logger.
func (r *Runtime) pullImage(ctx context.Context, ref string) error {
	rc, err := r.client.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(io.Discard, rc) // drain to completion
	return err
}
