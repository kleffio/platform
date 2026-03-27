// Package manual implements RuntimeAdapter for environments where the platform
// cannot manage containers automatically (no Docker socket, no k8s API).
// Deploy() returns a human-readable instruction instead of actually running
// anything. The operator runs the command manually and clicks Confirm in the UI.
package manual

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kleffio/platform/internal/shared/runtime"
)

// ManualInstruction is returned (wrapped in an error) by Deploy when the
// manual runtime is active. HTTP handlers detect this type and return a 202
// Accepted with the instruction text.
type ManualInstruction struct {
	Command string
}

func (m *ManualInstruction) Error() string {
	return "manual runtime: operator must run: " + m.Command
}

// Runtime implements runtime.RuntimeAdapter for the manual deployment mode.
type Runtime struct {
	// pluginAddrs maps plugin ID → "host:port".
	// Populated from PLUGIN_{ID}_ADDR environment variables at startup.
	pluginAddrs map[string]string
}

// New creates a ManualRuntime. addrs is a map of plugin-id → "host:port"
// for plugins already running in the environment.
func New(addrs map[string]string) *Runtime {
	if addrs == nil {
		addrs = make(map[string]string)
	}
	return &Runtime{pluginAddrs: addrs}
}

// ParseAddrsFromEnv reads PLUGIN_{PLUGIN_ID}_ADDR environment variables.
// Plugin IDs are normalised: "idp-auth0" → "IDP_AUTH0".
func ParseAddrsFromEnv() map[string]string {
	addrs := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		if strings.HasPrefix(key, "PLUGIN_") && strings.HasSuffix(key, "_ADDR") {
			// PLUGIN_IDP_AUTH0_ADDR → extract middle part
			mid := strings.TrimPrefix(key, "PLUGIN_")
			mid = strings.TrimSuffix(mid, "_ADDR")
			pluginID := strings.ToLower(strings.ReplaceAll(mid, "_", "-"))
			addrs[pluginID] = val
		}
	}
	return addrs
}

var _ runtime.RuntimeAdapter = (*Runtime)(nil)

// Deploy returns a ManualInstruction error so the HTTP handler can present
// the docker run command to the operator.
func (r *Runtime) Deploy(_ context.Context, spec runtime.ContainerSpec) error {
	var sb strings.Builder
	sb.WriteString("docker run -d \\\n")
	sb.WriteString(fmt.Sprintf("  --name %s \\\n", spec.ID))
	sb.WriteString("  --network kleff \\\n")
	for k, v := range spec.Env {
		sb.WriteString(fmt.Sprintf("  -e %s=%s \\\n", k, v))
	}
	for _, p := range spec.Ports {
		if p.HostPort != 0 {
			sb.WriteString(fmt.Sprintf("  -p %d:%d \\\n", p.HostPort, p.ContainerPort))
		}
	}
	sb.WriteString(fmt.Sprintf("  %s", spec.Image))
	return &ManualInstruction{Command: sb.String()}
}

func (r *Runtime) Remove(_ context.Context, _ string) error { return nil }
func (r *Runtime) Start(_ context.Context, _ string) error  { return nil }
func (r *Runtime) Stop(_ context.Context, _ string) error   { return nil }

func (r *Runtime) Status(_ context.Context, id string) (runtime.ContainerStatus, error) {
	if _, ok := r.pluginAddrs[id]; ok {
		return runtime.ContainerStatus{ID: id, State: runtime.StateRunning}, nil
	}
	return runtime.ContainerStatus{ID: id, State: runtime.StateUnknown}, nil
}

func (r *Runtime) Endpoint(_ context.Context, id string, port int) (string, error) {
	if addr, ok := r.pluginAddrs[id]; ok {
		return addr, nil
	}
	return fmt.Sprintf("%s:%d", id, port), nil
}

func (r *Runtime) Logs(_ context.Context, _ string, _ int) ([]string, error) {
	return []string{"manual runtime: logs not available"}, nil
}
